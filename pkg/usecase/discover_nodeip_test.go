package usecase_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/netip"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/scality/nodeip-discovery/pkg/domain"
	"github.com/scality/nodeip-discovery/pkg/infrastructure/ipextracter"
	"github.com/scality/nodeip-discovery/pkg/infrastructure/nodeannotater"
	"github.com/scality/nodeip-discovery/pkg/service"
	"github.com/scality/nodeip-discovery/pkg/usecase"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	nodeName          = "test-node"
	annotationRootKey = "loadbalancer.scality.com"
	applicationName   = "nodeip-discovery"
)

var _ = Describe("DiscoverNodeIP", func() {
	var (
		logger              *slog.Logger
		interfaceListerMock *service.MockInterfaceLister
		nodeAnnotater       service.NodeAnnotater
		ipExtracter         service.IPExtracter
		k8sClient           *fake.Clientset
		uc                  *usecase.DiscoverNodeIP
		ctx                 context.Context
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		interfaceListerMock = service.NewMockInterfaceLister(GinkgoT())
		ipExtracter = ipextracter.NewNetwork(
			logger,
			[]netip.Prefix{
				netip.MustParsePrefix("192.168.1.0/24"),
				netip.MustParsePrefix("192.168.2.0/24"),
				netip.MustParsePrefix("169.254.1.0/16"), // Link-Local Unicast for testing purpose
				netip.MustParsePrefix("224.0.0.0/24"),   // Multicast for testing purpose
			},
			[]netip.Addr{
				netip.MustParseAddr("192.168.1.100"),
			},
		)
		k8sClient = fake.NewSimpleClientset(&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
				Annotations: map[string]string{
					annotationRootKey: "192.168.0.10",
				},
			},
		})
		nodeAnnotater = nodeannotater.NewKubernetes(logger, k8sClient, nodeName, annotationRootKey, applicationName)
		uc = usecase.NewDiscoverNodeIP(logger, ipExtracter, nodeAnnotater, interfaceListerMock)
		ctx = context.Background()
	})

	Context("when matching IPs are found, including excluded IPs", func() {
		It("annotate correctly the Node, excluding the excluded IPs, and returns nil", func() {
			ifaces := []domain.IfaceAddrs{
				{
					Name:  "lo",
					Flags: net.FlagLoopback,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}},
					},
				},
				{
					Name:  "eth0",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 1, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "vip1",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 1, 100}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth1",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 2, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth2",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 3, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
			}
			interfaceListerMock.EXPECT().ListInterfaces().Return(ifaces, nil)

			err := uc.Execute(ctx)
			Expect(err).To(BeNil())

			node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(node.Annotations).To(HaveKeyWithValue(annotationRootKey, "192.168.1.10,192.168.2.10"))
		})
	})

	Context("when matching IPs are found but the annotation is already up-to-date", func() {
		It("leave the annotations as-is and returns nil", func() {
			By("updating, previously, manually, the node with the expected annotation")
			node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())

			node.Annotations[annotationRootKey] = "192.168.1.10,192.168.2.10"
			_, err = k8sClient.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
			Expect(err).To(BeNil())

			By("executing the usecase with the expected interfaces")
			ifaces := []domain.IfaceAddrs{
				{
					Name:  "lo",
					Flags: net.FlagLoopback,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}},
					},
				},
				{
					Name:  "eth0",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 1, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth1",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 2, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth2",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 3, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
			}
			interfaceListerMock.EXPECT().ListInterfaces().Return(ifaces, nil)

			err = uc.Execute(ctx)
			Expect(err).To(BeNil())

			By("checking that the annotation is still the same")
			node, err = k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(node.Annotations).To(HaveKeyWithValue(annotationRootKey, "192.168.1.10,192.168.2.10"))
		})
	})

	Context("when matching IPs are found but some are down", func() {
		It("skips the down interfaces, annotate the node with the up interfaces and returns nil", func() {
			ifaces := []domain.IfaceAddrs{
				{
					Name:  "lo",
					Flags: net.FlagLoopback,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}},
					},
				},
				{
					Name:  "eth0",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 1, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth1",
					Flags: 0,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 2, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth2",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 3, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
			}
			interfaceListerMock.EXPECT().ListInterfaces().Return(ifaces, nil)

			err := uc.Execute(ctx)
			Expect(err).To(BeNil())

			By("checking that the annotation is the expected one")
			node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(node.Annotations).To(HaveKeyWithValue(annotationRootKey, "192.168.1.10"))
		})
	})

	Context("when matching IPs are found but some are loopback", func() {
		It("skips the loopback interfaces, annotate the node with the non-loopback interfaces and returns nil", func() {
			ifaces := []domain.IfaceAddrs{
				{
					Name:  "lo",
					Flags: net.FlagLoopback,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}},
					},
				},
				{
					Name:  "eth0",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 1, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth1",
					Flags: net.FlagLoopback,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 2, 10}, Mask: net.IPMask{255, 255, 255, 255}},
					},
				},
				{
					Name:  "eth2",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 3, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
			}
			interfaceListerMock.EXPECT().ListInterfaces().Return(ifaces, nil)

			err := uc.Execute(ctx)
			Expect(err).To(BeNil())

			By("checking that the annotation is the expected one")
			node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(node.Annotations).To(HaveKeyWithValue(annotationRootKey, "192.168.1.10"))
		})
	})

	Context("when matching IPs are found but some are link-local", func() {
		It("skips the link-local interfaces, annotate the node with the non-link-local interfaces and returns nil", func() {
			ifaces := []domain.IfaceAddrs{
				{
					Name:  "lo",
					Flags: net.FlagLoopback,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}},
					},
				},
				{
					Name:  "eth0",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 1, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth1",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{169, 254, 1, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth2",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 3, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
			}
			interfaceListerMock.EXPECT().ListInterfaces().Return(ifaces, nil)

			err := uc.Execute(ctx)
			Expect(err).To(BeNil())

			By("checking that the annotation is the expected one")
			node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(node.Annotations).To(HaveKeyWithValue(annotationRootKey, "192.168.1.10"))
		})
	})

	Context("when matching IPs are found but some are link-local multicast", func() {
		It("skips the link-local multicast interfaces, annotate the node with the non-link-local multicast interfaces and returns nil", func() {
			ifaces := []domain.IfaceAddrs{
				{
					Name:  "lo",
					Flags: net.FlagLoopback,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{127, 0, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}},
					},
				},
				{
					Name:  "eth0",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 1, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth1",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{224, 0, 0, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
				{
					Name:  "eth2",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 3, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
			}
			interfaceListerMock.EXPECT().ListInterfaces().Return(ifaces, nil)

			err := uc.Execute(ctx)
			Expect(err).To(BeNil())

			By("checking that the annotation is the expected one")
			node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(node.Annotations).To(HaveKeyWithValue(annotationRootKey, "192.168.1.10"))
		})
	})

	Context("when no matching IPs are found", func() {
		It("should empty the annotation and returns nil", func() {
			ifaces := []domain.IfaceAddrs{
				{
					Name:  "eth0",
					Flags: net.FlagUp,
					Addresses: []net.Addr{
						&net.IPNet{IP: net.IP{192, 168, 3, 10}, Mask: net.IPMask{255, 255, 255, 0}},
					},
				},
			}
			interfaceListerMock.EXPECT().ListInterfaces().Return(ifaces, nil)

			err := uc.Execute(ctx)
			Expect(err).To(BeNil())

			node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			Expect(node.Annotations).To(HaveKeyWithValue(annotationRootKey, ""))
		})
	})
})
