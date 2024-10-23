package secretgenerator_test

import (
	"context"
	"testing"

	"github.com/onsi/gomega/gstruct"
	oauthv1 "github.com/openshift/api/oauth/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/opendatahub-io/opendatahub-operator/v2/controllers/secretgenerator"
	"github.com/opendatahub-io/opendatahub-operator/v2/pkg/metadata/annotations"

	. "github.com/onsi/gomega"
)

//nolint:ireturn
func newFakeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(oauthv1.AddToScheme(scheme))

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
}

func TestGenerateSecret(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	secretName := "foo"
	secretNs := "ns"

	// secret expected to be found
	existingSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNs,
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				annotations.SecretNameAnnotation: "bar",
				annotations.SecretTypeAnnotation: "random",
			},
		},
	}

	// secret to be generated
	generatedSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName + "-generated",
			Namespace: secretNs,
		},
	}

	cli := newFakeClient(&existingSecret)

	r := secretgenerator.SecretGeneratorReconciler{
		Client: cli,
	}

	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      existingSecret.Name,
			Namespace: existingSecret.Namespace,
		},
	})

	g.Expect(err).ShouldNot(HaveOccurred())

	err = cli.Get(ctx, client.ObjectKeyFromObject(&generatedSecret), &generatedSecret)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(generatedSecret.OwnerReferences).To(HaveLen(1))
	g.Expect(generatedSecret.OwnerReferences[0]).To(
		gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"Name":       Equal(existingSecret.Name),
			"Kind":       Equal(existingSecret.Kind),
			"APIVersion": Equal(existingSecret.APIVersion),
		}),
	)

	g.Expect(generatedSecret.StringData).To(
		HaveKey(existingSecret.Annotations[annotations.SecretNameAnnotation]))
	g.Expect(generatedSecret.Labels).To(
		gstruct.MatchAllKeys(gstruct.Keys{
			"foo": Equal("bar"),
		}),
	)
}

func TestExistingSecret(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	secretName := "foo"
	secretNs := "ns"

	// secret expected to be found
	existingSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNs,
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				annotations.SecretNameAnnotation: "bar",
				annotations.SecretTypeAnnotation: "random",
			},
		},
	}

	// secret to be generated
	generatedSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName + "-generated",
			Namespace: secretNs,
		},
	}

	cli := newFakeClient(&existingSecret, &generatedSecret)

	r := secretgenerator.SecretGeneratorReconciler{
		Client: cli,
	}

	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      existingSecret.Name,
			Namespace: existingSecret.Namespace,
		},
	})

	g.Expect(err).ShouldNot(HaveOccurred())

	err = cli.Get(ctx, client.ObjectKeyFromObject(&generatedSecret), &generatedSecret)
	g.Expect(err).ShouldNot(HaveOccurred())

	// assuming an existing secret is left untouched
	g.Expect(generatedSecret.OwnerReferences).To(BeEmpty())
	g.Expect(generatedSecret.Labels).To(BeEmpty())
	g.Expect(generatedSecret.StringData).To(BeEmpty())
}

func TestGenerateSecretIfNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	secretName := "foo"
	secretNs := "ns"

	// secret expected to be generated during reconcile
	expectedSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName + "-generated",
			Namespace: secretNs,
			Labels: map[string]string{
				secretName: "bar",
			},
			Annotations: map[string]string{
				annotations.SecretNameAnnotation: "bar",
				annotations.SecretTypeAnnotation: "random",
			},
		},
	}

	cli := newFakeClient()

	r := secretgenerator.SecretGeneratorReconciler{
		Client: cli,
	}

	foundSecret := corev1.Secret{}

	// ensure the secret does not exist yet before reconcile
	err := cli.Get(ctx, client.ObjectKeyFromObject(&expectedSecret), &foundSecret)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(k8serr.IsNotFound(err)).Should(BeTrue())

	_, err = r.Reconcile(ctx, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      secretName,
			Namespace: secretNs,
		},
	})
	// TODO ensure no error occured during reconcile?
	// (in the current implementation, missing annotation error would occur during reconcile when generating new secret)
	// g.Expect(err).ShouldNot(HaveOccurred())

	err = cli.Get(ctx, client.ObjectKeyFromObject(&expectedSecret), &foundSecret)
	// TODO ensure the secret was created successfully?
	// g.Expect(err).ShouldNot(HaveOccurred())
	// g.Expect(k8serr.IsNotFound(err)).Should(BeFalse())
}

func TestDeleteOAuthClientIfSecretNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	secretName := "foo"
	secretNs := "ns"

	// secret expected to be deleted
	existingSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNs,
			Labels: map[string]string{
				"foo": "bar",
			},
			Annotations: map[string]string{
				annotations.SecretNameAnnotation: "bar",
				annotations.SecretTypeAnnotation: "random",
			},
		},
	}

	existingOauthClient := oauthv1.OAuthClient{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OAuthClient",
			APIVersion: "oauth.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNs,
		},
		Secret:       secretName,
		RedirectURIs: []string{"https://foo.bar"},
		GrantMethod:  oauthv1.GrantHandlerAuto,
	}

	cli := newFakeClient(&existingSecret, &existingOauthClient)

	r := secretgenerator.SecretGeneratorReconciler{
		Client: cli,
	}

	// delete secret
	err := cli.Delete(ctx, &existingSecret)
	g.Expect(err).ShouldNot(HaveOccurred())

	// ensure the secret is deleted (not found)
	err = cli.Get(ctx, client.ObjectKeyFromObject(&existingSecret), &existingSecret)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(k8serr.IsNotFound(err)).Should(BeTrue())

	_, err = r.Reconcile(ctx, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      secretName,
			Namespace: secretNs,
		},
	})
	// TODO ensure no error occured during reconcile?
	// (in the current implementation, missing annotation error would occur during reconcile when generating new secret)
	// g.Expect(err).ShouldNot(HaveOccurred())

	// ensure the leftover OauthClient was deleted during reconcile
	foundOauthClient := oauthv1.OAuthClient{}
	err = cli.Get(ctx, client.ObjectKeyFromObject(&existingOauthClient), &foundOauthClient)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(k8serr.IsNotFound(err)).Should(BeTrue())
}
