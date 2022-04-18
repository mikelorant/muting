package mutationconfig

import (
	"bytes"
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func CreateClient() *kubernetes.Clientset {
	config := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("Failed to configure client.")
	}

	return kubeClient
}

func GenerateMutationConfig(mutationCfgName string, webhookNamespace string, webhookService string, caCert *bytes.Buffer) (mutateConfig *admissionregistrationv1.MutatingWebhookConfiguration) {
	path := "/mutate"
	fail := admissionregistrationv1.Fail
	sideEffect := admissionregistrationv1.SideEffectClassNone

	service := &admissionregistrationv1.ServiceReference{
		Name:      webhookService,
		Namespace: webhookNamespace,
		Path:      &path,
	}

	// Used for local debugging.
	// url := "https://host.minikube.internal:6883"+path

	mutateConfig = &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: mutationCfgName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:                    fmt.Sprint(webhookService, ".", webhookNamespace, ".svc.cluster.local"),
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffect,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caCert.Bytes(), // CA bundle created earlier
				Service:  service,
				// URL: &url,
			},
			Rules: []admissionregistrationv1.RuleWithOperations{{
				Operations: []admissionregistrationv1.OperationType{
					admissionregistrationv1.Create,
					admissionregistrationv1.Update,
				},
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{"networking.k8s.io"},
					APIVersions: []string{"v1"},
					Resources:   []string{"ingresses"},
				},
			}},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					(webhookService): "enabled",
				},
			},
			FailurePolicy: &fail,
		}},
	}

	return mutateConfig
}

func ApplyMutationConfig(client *kubernetes.Clientset, mutationCfgName string, mutateConfig *admissionregistrationv1.MutatingWebhookConfiguration) error {
	existingConfig, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.TODO(), mutationCfgName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), mutateConfig, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		mutateConfig.ObjectMeta.ResourceVersion = existingConfig.ObjectMeta.ResourceVersion
		if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.TODO(), mutateConfig, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	return nil
}
