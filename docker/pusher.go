package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	fn "knative.dev/kn-plugin-func"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Opt func(*Pusher) error

type Credentials struct {
	Username string
	Password string
}

type CredentialsProvider func(ctx context.Context, registry string) (Credentials, error)

// Pusher of images from local to remote registry.
type Pusher struct {
	// Verbose logging.
	Verbose             bool
	credentialsProvider CredentialsProvider
	progressListener    fn.ProgressListener
	transport           http.RoundTripper
}

func WithCredentialsProvider(cp CredentialsProvider) Opt {
	return func(p *Pusher) error {
		p.credentialsProvider = cp
		return nil
	}
}

func WithProgressListener(pl fn.ProgressListener) Opt {
	return func(p *Pusher) error {
		p.progressListener = pl
		return nil
	}
}

func WithTransport(transport http.RoundTripper) Opt {
	return func(pusher *Pusher) error {
		pusher.transport = transport
		return nil
	}
}

func EmptyCredentialsProvider(ctx context.Context, registry string) (Credentials, error) {
	return Credentials{}, nil
}

// NewPusher creates an instance of a docker-based image pusher.
func NewPusher(opts ...Opt) (*Pusher, error) {
	result := &Pusher{
		Verbose:             false,
		credentialsProvider: EmptyCredentialsProvider,
		progressListener:    &fn.NoopProgressListener{},
		transport:           http.DefaultTransport,
	}
	for _, opt := range opts {
		err := opt(result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func GetRegistry(image_url string) (string, error) {
	var registry string
	parts := strings.Split(image_url, "/")
	switch {
	case len(parts) == 2:
		registry = fn.DefaultRegistry
	case len(parts) >= 3:
		registry = parts[0]
	default:
		return "", fmt.Errorf("failed to parse image name: %q", image_url)
	}

	return registry, nil
}

// Push the image of the Function.
func (n *Pusher) Push(ctx context.Context, f fn.Function) (digest string, err error) {

	var output io.Writer

	if n.Verbose {
		output = os.Stderr
	} else {
		output = io.Discard
	}

	if f.Image == "" {
		return "", errors.New("Function has no associated image.  Has it been built?")
	}

	registry, err := GetRegistry(f.Image)
	if err != nil {
		return "", err
	}

	n.progressListener.Stopping()
	credentials, err := n.credentialsProvider(ctx, registry)
	if err != nil {
		return "", fmt.Errorf("failed to get credentials: %w", err)
	}
	n.progressListener.Increment("Pushing function image to the registry")

	// if the registry is not cluster private do push directly from daemon
	if _, err = net.DefaultResolver.LookupHost(ctx, registry); err == nil {
		return n.daemonPush(ctx, f, credentials, output)
	}

	// push with custom transport to be able to push into cluster private registries
	return n.push(ctx, f, credentials, output)
}

func (n *Pusher) daemonPush(ctx context.Context, f fn.Function, credentials Credentials, output io.Writer) (digest string, err error) {
	cli, _, err := NewClient(client.DefaultDockerHost)
	if err != nil {
		return "", fmt.Errorf("failed to create docker api client: %w", err)
	}
	defer cli.Close()

	authConfig := types.AuthConfig{
		Username: credentials.Username,
		Password: credentials.Password,
	}

	b, err := json.Marshal(&authConfig)
	if err != nil {
		return "", err
	}

	opts := types.ImagePushOptions{RegistryAuth: base64.StdEncoding.EncodeToString(b)}

	r, err := cli.ImagePush(ctx, f.Image, opts)
	if err != nil {
		return "", fmt.Errorf("failed to push the image: %w", err)
	}
	defer r.Close()

	var outBuff bytes.Buffer
	output = io.MultiWriter(&outBuff, output)

	decoder := json.NewDecoder(r)
	li := logItem{}
	for {
		err = decoder.Decode(&li)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if li.Error != "" {
			return "", errors.New(li.ErrorDetail.Message)
		}
		if li.Id != "" {
			fmt.Fprintf(output, "%s: ", li.Id)
		}
		var percent int
		if li.ProgressDetail.Total == 0 {
			percent = 100
		} else {
			percent = (li.ProgressDetail.Current * 100) / li.ProgressDetail.Total
		}
		fmt.Fprintf(output, "%s (%d%%)\n", li.Status, percent)
	}

	digest = ParseDigest(outBuff.String())

	return digest, nil
}

var digestRE = regexp.MustCompile(`digest:\s+(sha256:\w{64})`)

// ParseDigest tries to parse the last line from the output, which holds the pushed image digest
// The output should contain line like this:
// latest: digest: sha256:a278a91112d17f8bde6b5f802a3317c7c752cf88078dae6f4b5a0784deb81782 size: 2613
func ParseDigest(output string) string {
	match := digestRE.FindStringSubmatch(output)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

type errorDetail struct {
	Message string `json:"message"`
}

type progressDetail struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

type logItem struct {
	Id             string         `json:"id"`
	Status         string         `json:"status"`
	Error          string         `json:"error"`
	ErrorDetail    errorDetail    `json:"errorDetail"`
	Progress       string         `json:"progress"`
	ProgressDetail progressDetail `json:"progressDetail"`
}

func (n *Pusher) push(ctx context.Context, f fn.Function, credentials Credentials, output io.Writer) (digest string, err error) {
	auth := &authn.Basic{
		Username: credentials.Username,
		Password: credentials.Password,
	}

	ref, err := name.ParseReference(f.Image)
	if err != nil {
		return "", err
	}

	dockerClient, _, err := NewClient(client.DefaultDockerHost)
	if err != nil {
		return "", fmt.Errorf("failed to create docker api client: %w", err)
	}
	defer dockerClient.Close()

	img, err := daemon.Image(ref,
		daemon.WithContext(ctx),
		daemon.WithClient(dockerClient))
	if err != nil {
		return "", err
	}

	progressChannel := make(chan v1.Update, 1024)
	errChan := make(chan error)
	go func() {
		defer fmt.Fprint(output, "\n")

		for progress := range progressChannel {
			if progress.Error != nil {
				errChan <- progress.Error
				return
			}
			fmt.Fprintf(output, "\rprogress: %d%%", progress.Complete*100/progress.Total)
		}

		errChan <- nil
	}()

	err = remote.Write(ref, img,
		remote.WithAuth(auth),
		remote.WithProgress(progressChannel),
		remote.WithTransport(n.transport),
		remote.WithContext(ctx))
	if err != nil {
		return "", err
	}
	err = <-errChan
	if err != nil {
		return "", err
	}

	hash, err := img.Digest()
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}
