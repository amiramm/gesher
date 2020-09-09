/*
Copyright 2020 Redis Labs Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package admission_tester_test

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/redislabs/gesher/integration-tests/common"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAdmissionTester(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AdmissionTester Suite")
}

var (
	sa                 *corev1.ServiceAccount
	role               *rbacv1beta1.Role
	roleBinding        *rbacv1beta1.RoleBinding
	clusterRole        *rbacv1beta1.ClusterRole
	clusterRoleBinding *rbacv1beta1.ClusterRoleBinding

	webhook *v1beta1.ValidatingWebhookConfiguration
	service *corev1.Service
	secret  *corev1.Secret
	deploy  *appsv1.Deployment

	kubeClient client.Client
)

var _ = BeforeSuite(func() {
	var err error
	kubeClient, _, err = common.GetClient()
	Expect(err).To(Succeed())

	sa = common.LoadServiceAccount()
	role = common.LoadRole()
	roleBinding = common.LoadRoleBinding()
	clusterRole = common.LoadClusterRole()
	clusterRoleBinding = common.LoadClusterRoleBinding()

	service = common.LoadTestService()
	deploy = common.LoadAdmissionDeploy()

	var s corev1.Secret
	Expect(kubeClient.Get(context.TODO(), types.NamespacedName{Name: "admission-test", Namespace: common.Namespace}, &s)).To(Succeed())
	secret = &s

	webhook = loadWebhook(secret)
})

var _ = AfterSuite(func() {
	if webhook != nil {
		Expect(kubeClient.Delete(context.TODO(), webhook)).To(Succeed())
		webhook = nil
	}
	if service != nil {
		Expect(kubeClient.Delete(context.TODO(), service)).To(Succeed())
		service = nil
	}
	if deploy != nil {
		Expect(kubeClient.Delete(context.TODO(), deploy)).To(Succeed())
		deploy = nil
	}
	if secret != nil {
		Expect(kubeClient.Delete(context.TODO(), secret)).To(Succeed())
		secret = nil
	}
	if sa != nil {
		Expect(kubeClient.Delete(context.TODO(), sa)).To(Succeed())
		sa = nil
	}
	if clusterRole != nil {
		Expect(kubeClient.Delete(context.TODO(), clusterRole)).To(Succeed())
		clusterRole = nil
	}
	if clusterRoleBinding != nil {
		Expect(kubeClient.Delete(context.TODO(), clusterRoleBinding)).To(Succeed())
		clusterRoleBinding = nil
	}
	if role != nil {
		Expect(kubeClient.Delete(context.TODO(), role)).To(Succeed())
		role = nil
	}
	if roleBinding != nil {
		Expect(kubeClient.Delete(context.TODO(), roleBinding)).To(Succeed())
		roleBinding = nil
	}
})

func loadWebhook(s *corev1.Secret) *v1beta1.ValidatingWebhookConfiguration {
	By("Read and Load Webhook")
	
	path := "/admission"
	failurePolicy := v1beta1.Fail
	scope         := v1beta1.AllScopes

	webhook := &v1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "admission-test",
		},
		Webhooks:   []v1beta1.ValidatingWebhook{
			{
				Name:                    "test.admission.gesher",
				ClientConfig:            v1beta1.WebhookClientConfig{
					Service:  &v1beta1.ServiceReference{
						Name: "admission-test",
						Namespace: common.Namespace,
						Path: &path, 
					},
					CABundle: s.Data["cert"],
				},
				Rules:                   []v1beta1.RuleWithOperations{
					{
						Operations: []v1beta1.OperationType{v1beta1.Create},
						Rule:       v1beta1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"namespaces"},
							Scope: &scope,
						},
					},
				},
				FailurePolicy: &failurePolicy,
			},
		},
	}

	Expect(kubeClient.Create(context.TODO(), webhook)).To(Succeed())

	return webhook
}
