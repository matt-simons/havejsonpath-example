package main

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("HaveJSONPathMatcher Examples", func() {

	SetDefaultEventuallyTimeout(time.Minute)

	Describe("Asserting on a condition", func() {

		var k Komega
		var c client.Client
		var deployment *appsv1.Deployment

		BeforeEach(func() {
			var err error
			c, err = client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme})
			Expect(err).ShouldNot(HaveOccurred())
			k = New(c)
			appLabel := map[string]string{"app": "my-deployment"}
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: appLabel,
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: appLabel,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image:   "busybox:latest",
									Name:    "test",
									Command: []string{"/bin/sh"},
									Args:    []string{"-c", "sleep 60"},
								},
								{
									Image:   "busybox:latest",
									Name:    "sidecar",
									Command: []string{"/bin/sh"},
									Args:    []string{"-c", "sleep 60"},
								},
							},
						},
					},
				},
			}
		})

		Context("Without HaveJSONPath", Ordered, func() {
			BeforeAll(func(ctx SpecContext) {
				Eventually(c.Create).WithContext(ctx).WithArguments(deployment).Should(Succeed())
			}, NodeTimeout(5*time.Second))

			It("Should assert on replicas", func(ctx SpecContext) {
				k := k.WithContext(ctx)
				Eventually(k.Object(deployment)).Should(HaveField("Status.Replicas", BeEquivalentTo(1)))
			}, SpecTimeout(time.Minute))

			It("Should assert on conditions", func(ctx SpecContext) {
				k := k.WithContext(ctx)
				id := func(element interface{}) string {
					return string(element.(appsv1.DeploymentCondition).Type)
				}
				Eventually(k.Object(deployment)).Should(HaveField("Status.Conditions",
					MatchElements(id, IgnoreExtras, Elements{
						"Available": MatchFields(IgnoreExtras, Fields{
							"Reason": Equal("MinimumReplicasAvailable"),
						}),
					}),
				))
			}, SpecTimeout(time.Minute))

			AfterAll(func(ctx SpecContext) {
				Eventually(c.Delete).WithContext(ctx).WithArguments(deployment).Should(Succeed())
			}, NodeTimeout(time.Minute))

		})

		Context("With HaveJSONPath", Ordered, func() {
			BeforeAll(func(ctx SpecContext) {
				Eventually(c.Create).WithContext(ctx).WithArguments(deployment).Should(Succeed())
			}, NodeTimeout(5*time.Second))

			It("Should assert on replicas", func(ctx SpecContext) {
				k := k.WithContext(ctx)
				Eventually(k.Object(deployment)).Should(HaveJSONPath("{.status.replicas}", BeEquivalentTo(1)))
			}, SpecTimeout(time.Minute))

			It("Should assert on conditions", func(ctx SpecContext) {
				k := k.WithContext(ctx)
				Eventually(k.Object(deployment)).Should(HaveJSONPath(
					`{.status.conditions[?(@.type=="Available")].reason}`, Equal("MinimumReplicasAvailable")),
				)
			}, SpecTimeout(time.Minute))

			AfterAll(func(ctx SpecContext) {
				Eventually(c.Delete).WithContext(ctx).WithArguments(deployment).Should(Succeed())
			}, NodeTimeout(time.Minute))
		})
	})
})

func TestWithJSONPath(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestWithJSONPath")
}
