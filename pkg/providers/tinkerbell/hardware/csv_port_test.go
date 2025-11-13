package hardware_test

import (
	"strings"
	"testing"

	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestCSVReader_BMCPortOptional_NoColumn(t *testing.T) {
	g := gomega.NewWithT(t)

	// CSV without bmc_port column at all - should work fine (backward compatible)
	// This is the exact format existing users have been using
	csvData := `hostname,bmc_ip,bmc_username,bmc_password,mac,ip_address,netmask,gateway,nameservers,labels,disk
worker1,192.168.0.10,Admin,admin,00:00:00:00:00:01,10.10.10.10,255.255.255.0,10.10.10.1,1.1.1.1,type=cp,/dev/sda`

	csvReader, err := hardware.NewCSVReader(strings.NewReader(csvData), nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := csvReader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine.BMCPort).To(gomega.Equal(0), "BMCPort should default to 0 when column is not present")
	g.Expect(machine.BMCIPAddress).To(gomega.Equal("192.168.0.10"))
}

func TestCSVReader_BMCPortOptional_EmptyValue(t *testing.T) {
	g := gomega.NewWithT(t)

	// CSV with bmc_port column but empty value - should work fine
	csvData := `hostname,bmc_ip,bmc_username,bmc_password,bmc_port,mac,ip_address,netmask,gateway,nameservers,labels,disk
worker1,192.168.0.10,Admin,admin,,00:00:00:00:00:01,10.10.10.10,255.255.255.0,10.10.10.1,1.1.1.1,type=cp,/dev/sda`

	csvReader, err := hardware.NewCSVReader(strings.NewReader(csvData), nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := csvReader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine.BMCPort).To(gomega.Equal(0), "BMCPort should default to 0 when value is empty")
	g.Expect(machine.BMCIPAddress).To(gomega.Equal("192.168.0.10"))
}

func TestCSVReader_BMCPortOptional_WithValue(t *testing.T) {
	g := gomega.NewWithT(t)

	// CSV with bmc_port column and specific value - should use that value
	csvData := `hostname,bmc_ip,bmc_username,bmc_password,bmc_port,mac,ip_address,netmask,gateway,nameservers,labels,disk
worker1,192.168.0.10,Admin,admin,6230,00:00:00:00:00:01,10.10.10.10,255.255.255.0,10.10.10.1,1.1.1.1,type=cp,/dev/sda`

	csvReader, err := hardware.NewCSVReader(strings.NewReader(csvData), nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	machine, err := csvReader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine.BMCPort).To(gomega.Equal(6230), "BMCPort should use specified value")
	g.Expect(machine.BMCIPAddress).To(gomega.Equal("192.168.0.10"))
}

func TestCSVReader_BMCPortOptional_MixedRows(t *testing.T) {
	g := gomega.NewWithT(t)

	// CSV with mixed scenarios - some machines with port, some without
	csvData := `hostname,bmc_ip,bmc_username,bmc_password,bmc_port,mac,ip_address,netmask,gateway,nameservers,labels,disk
worker1,192.168.0.10,Admin,admin,6230,00:00:00:00:00:01,10.10.10.10,255.255.255.0,10.10.10.1,1.1.1.1,type=cp,/dev/sda
worker2,192.168.0.10,Admin,admin,,00:00:00:00:00:02,10.10.10.11,255.255.255.0,10.10.10.1,1.1.1.1,type=worker,/dev/sda
worker3,192.168.0.10,Admin,admin,6231,00:00:00:00:00:03,10.10.10.12,255.255.255.0,10.10.10.1,1.1.1.1,type=worker,/dev/sda`

	csvReader, err := hardware.NewCSVReader(strings.NewReader(csvData), nil)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// First machine has port 6230
	machine1, err := csvReader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine1.BMCPort).To(gomega.Equal(6230))

	// Second machine has empty port (defaults to 0)
	machine2, err := csvReader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine2.BMCPort).To(gomega.Equal(0))

	// Third machine has port 6231
	machine3, err := csvReader.Read()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(machine3.BMCPort).To(gomega.Equal(6231))
}

func TestBMCCatalogueWriter_PortDefaultBehavior(t *testing.T) {
	testCases := []struct {
		name         string
		bmcPort      int
		expectedPort int
	}{
		{
			name:         "Port not specified (0) - Rufio uses default 623",
			bmcPort:      0,
			expectedPort: 0, // 0 means Rufio will use its default (623)
		},
		{
			name:         "Custom port specified",
			bmcPort:      6230,
			expectedPort: 6230,
		},
		{
			name:         "Standard IPMI port",
			bmcPort:      623,
			expectedPort: 623,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			catalogue := hardware.NewCatalogue()
			writer := hardware.NewBMCCatalogueWriter(catalogue)

			machine := NewValidMachine()
			machine.BMCPort = tc.bmcPort

			err := writer.Write(machine)
			g.Expect(err).To(gomega.Succeed())

			bmcs := catalogue.AllBMCs()
			g.Expect(bmcs).To(gomega.HaveLen(1))
			g.Expect(bmcs[0].Spec.Connection.Port).To(gomega.Equal(tc.expectedPort))
		})
	}
}

