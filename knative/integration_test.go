//go:build integration
// +build integration

package knative_test

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	v1 "knative.dev/pkg/apis/duck/v1"

	fn "knative.dev/func"
	"knative.dev/func/k8s"
	"knative.dev/func/knative"
)

// Basic happy path test of deploy->describe->list->re-deploy->delete.
func TestIntegration(t *testing.T) {
	var err error
	functionName := "fn-testing"

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	t.Cleanup(cancel)

	cliSet, err := k8s.NewKubernetesClientset()
	if err != nil {
		t.Fatal(err)
	}

	namespace := "knative-integration-test-ns-" + rand.String(5)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{},
	}
	_, err = cliSet.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cliSet.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{}) })
	t.Log("created namespace: ", namespace)

	secret := "credentials-secret"
	sc := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secret,
		},
		Data: map[string][]byte{
			"FUNC_TEST_SC_A": []byte("A"),
			"FUNC_TEST_SC_B": []byte("B"),
		},
		StringData: nil,
		Type:       corev1.SecretTypeOpaque,
	}

	sc, err = cliSet.CoreV1().Secrets(namespace).Create(ctx, sc, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	configMap := "testing-config-map"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMap,
		},
		Data: map[string]string{"FUNC_TEST_CM_A": "1"},
	}
	cm, err = cliSet.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	trigger := "testing-trigger"
	tr := &eventingv1.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Name: trigger,
		},
		Spec: eventingv1.TriggerSpec{
			Broker: "testing-broker",
			Subscriber: v1.Destination{Ref: &v1.KReference{
				Kind:       "Service",
				Namespace:  namespace,
				Name:       functionName,
				APIVersion: "serving.knative.dev/v1",
			}},
			Filter: &eventingv1.TriggerFilter{
				Attributes: map[string]string{
					"source": "test-event-source",
					"type":   "test-event-type",
				},
			},
		},
	}

	eventingClient, err := knative.NewEventingClient(namespace)
	if err != nil {
		t.Fatal(err)
	}
	err = eventingClient.CreateTrigger(ctx, tr)
	if err != nil {
		t.Fatal(err)
	}

	minScale := int64(2)
	maxScale := int64(100)

	now := time.Now()
	function := fn.Function{
		SpecVersion: "SNAPSHOT",
		Root:        "/non/existent",
		Name:        functionName,
		Runtime:     "blub",
		Template:    "cloudevents",
		// Basic "echo" Go Function that in addition:
		// * prints environment variables starting which name starts with FUNC_TEST to stderr,
		// * lists files under /etc/cm and /etc/sc stderr.
		Image:       "quay.io/mvasek/func-test-service",
		ImageDigest: "sha256:69251ac335693c4d5a503e9f8a829a25b93bff6e6bddbff5b78971fe668dc71a",
		Created:     now,
		Deploy: fn.DeploySpec{
			Namespace: namespace,
			Labels:    []fn.Label{{Key: ptr("my-label"), Value: ptr("my-label-value")}},
			Options: fn.Options{
				Scale: &fn.ScaleOptions{
					Min: &minScale,
					Max: &maxScale,
				},
			},
		},
		Run: fn.RunSpec{
			Envs: []fn.Env{
				{Name: ptr("FUNC_TEST_VAR"), Value: ptr("nbusr123")},
				{Name: ptr("FUNC_TEST_SC_A"), Value: ptr("{{ secret: " + secret + ":FUNC_TEST_SC_A }}")},
				{Value: ptr("{{configMap:" + configMap + "}}")},
			},
			Volumes: []fn.Volume{
				{Secret: ptr(secret), Path: ptr("/etc/sc")},
				{ConfigMap: ptr(configMap), Path: ptr("/etc/cm")},
			},
		},
	}

	var buff = &knative.SynchronizedBuffer{}
	go func() {
		_ = knative.GetKServiceLogs(ctx, namespace, functionName, function.ImageWithDigest(), &now, buff)
	}()

	deployer := knative.NewDeployer(knative.WithDeployerNamespace(namespace), knative.WithDeployerVerbose(false))

	depRes, err := deployer.Deploy(ctx, function)
	if err != nil {
		t.Fatal(err)
	}

	outStr := buff.String()
	t.Logf("deploy result: %+v", depRes)
	t.Log("function output:\n" + outStr)

	if strings.Count(outStr, "starting app") < int(minScale) {
		t.Errorf("application should be scaled at least to %d pods", minScale)
	}

	// verify that environment variables and volumes works
	if !strings.Contains(outStr, "FUNC_TEST_VAR=nbusr123") {
		t.Error("plain environment variable was not propagated")
	}
	if !strings.Contains(outStr, "FUNC_TEST_SC_A=A") {
		t.Error("environment variables from secret was not propagated")
	}
	if strings.Contains(outStr, "FUNC_TEST_SC_B=") {
		t.Error("environment variables from secret was propagated but should have not been")
	}
	if !strings.Contains(outStr, "FUNC_TEST_CM_A=1") {
		t.Error("environment variable from config-map was not propagated")
	}
	if !strings.Contains(outStr, "/etc/sc/FUNC_TEST_SC_A") || !strings.Contains(outStr, "/etc/sc/FUNC_TEST_SC_A") {
		t.Error("secret was not mounted")
	}
	if !strings.Contains(outStr, "/etc/cm/FUNC_TEST_CM_A") {
		t.Error("config-map was not mounted")
	}

	describer := knative.NewDescriber(namespace, false)
	instance, err := describer.Describe(ctx, functionName)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("instance: %+v", instance)

	// verify that trigger info is included in describe output
	if len(instance.Subscriptions) != 1 {
		t.Error("exactly one subscription is expected")
	} else {
		if instance.Subscriptions[0].Broker != "testing-broker" {
			t.Fatal("bad broker")
		}
		if instance.Subscriptions[0].Source != "test-event-source" {
			t.Fatal("bad source")
		}
		if instance.Subscriptions[0].Type != "test-event-type" {
			t.Fatal("bad type")
		}
	}

	lister := knative.NewLister(namespace, false)
	list, err := lister.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("functions list: %+v", list)

	if len(list) != 1 {
		t.Errorf("expected exactly one functions but got: %d", len(list))
	} else {
		if list[0].URL != instance.Route {
			t.Error("URL mismatch")
		}
	}

	buff.Reset()
	t.Setenv("LOCAL_ENV_TO_DEPLOY", "iddqd")
	function.Run.Envs = []fn.Env{
		{Name: ptr("FUNC_TEST_VAR"), Value: ptr("{{ env:LOCAL_ENV_TO_DEPLOY }}")},
		{Value: ptr("{{ secret: " + secret + " }}")},
		{Name: ptr("FUNC_TEST_CM_A_ALIASED"), Value: ptr("{{configMap:" + configMap + ":FUNC_TEST_CM_A}}")},
	}
	depRes, err = deployer.Deploy(ctx, function)
	if err != nil {
		t.Fatal(err)
	}
	outStr = buff.String()
	t.Log("function output:\n" + outStr)

	// verify that environment variables has been changed by re-deploy
	if strings.Contains(outStr, "FUNC_TEST_CM_A=") {
		t.Error("environment variables from previous deployment was not removed")
	}
	if !strings.Contains(outStr, "FUNC_TEST_SC_A=A") || !strings.Contains(outStr, "FUNC_TEST_SC_B=B") {
		t.Error("environment variables were not imported from secret")
	}
	if !strings.Contains(outStr, "FUNC_TEST_VAR=iddqd") {
		t.Error("environment variable was not set from local environment variable")
	}
	if !strings.Contains(outStr, "FUNC_TEST_CM_A_ALIASED=1") {
		t.Error("environment variable was not set from config-map")
	}

	remover := knative.NewRemover(namespace, false)
	err = remover.Remove(ctx, functionName)
	if err != nil {
		t.Fatal(err)
	}

	list, err = lister.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 0 {
		t.Errorf("expected exactly zero functions but got: %d", len(list))
	}
}

func ptr[T interface{}](s T) *T {
	return &s
}
