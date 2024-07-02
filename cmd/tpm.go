package cmd

import (
	"encoding/base64"
	"os"

	"github.com/google/go-attestation/attest"

	"github.com/spf13/cobra"
)

var TPMImportEkCert string
var TPMBase64 bool
var TPMVerify bool
var TPMEventLog bool

func init() {

	tpmCmd.PersistentFlags().StringVar(&TPMImportEkCert, "import-ek-cert", "", "Import the TPM Endorsement Key (EK) into the Certificate Authority")
	tpmCmd.PersistentFlags().BoolVar(&TPMBase64, "base64", true, "Flag indicating if the certificate is base64 encoded")
	tpmCmd.PersistentFlags().BoolVar(&TPMVerify, "verify", true, "Flag indicating if the certificate should be verified by the Certificate Authority")
	tpmCmd.PersistentFlags().BoolVar(&TPMEventLog, "event-log", false, "Shows the TPM event log")

	rootCmd.AddCommand(tpmCmd)
}

var tpmCmd = &cobra.Command{
	Use:   "tpm",
	Short: "Trusted Platform Module",
	Long: `The Trusted Platform Module (TPM) technology is designed to provide hardware-based,
		   security-related functions. A TPM chip is a secure crypto-processor that is designed
		   to carry out cryptographic operations. The chip includes multiple physical security
		   mechanisms to make it tamper-resistant, and malicious software is unable to tamper
		   with the security functions of the TPM. Some of the advantages of using TPM technology
		   are:

		   - Generate, store, and limit the use of cryptographic keys.
		   - Use it for device authentication by using the TPM's unique RSA key, which is burned into the chip.
		   - Help ensure platform integrity by taking and storing security measurements of the boot process.

		   For more information 
		   https://trustedcomputinggroup.org/about/what-is-a-trusted-platform-module-tpm/
		   https://learn.microsoft.com/en-us/windows/security/hardware-security/tpm/trusted-platform-module-overview
		   https://www.intel.com/content/www/us/en/business/enterprise-computers/resources/trusted-platform-module.html
		   `,
	Run: func(cmd *cobra.Command, args []string) {

		cn := "ek-tpm"

		// --import-ek-cert cn
		if TPMImportEkCert != "" {

			bytes, err := os.ReadFile(TPMImportEkCert)
			if err != nil {
				App.Logger.Fatal(err)
			}

			decoded := make([]byte, len(bytes))
			if TPMBase64 {
				n, err := base64.StdEncoding.Decode(decoded, bytes)
				if err != nil {
					App.Logger.Fatal(err)
				}
				if n == 0 {
					App.Logger.Fatal("invalid format: unable to base64 decode certificate")
				}
			} else {
				decoded = bytes
			}

			ekCert, err := attest.ParseEKCertificate(decoded)
			if err != nil {
				App.Logger.Fatal(err)
			}

			if err := App.CA.ImportCN(cn, ekCert); err != nil {
				App.Logger.Fatal(err)
			}

			if TPMVerify {
				valid, err := App.CA.Verify(ekCert, &cn)
				if err != nil {
					App.Logger.Fatal(err)
				}
				if !valid {
					App.Logger.Fatal("failed to validate EK certificate")
				}
			}

			App.Logger.Info("EK certificate successfully imported")
			os.Exit(0)
		}

		// --event-log
		if TPMEventLog {

			// measurementLog, err := os.ReadFile("/sys/kernel/security/tpm0/binary_bios_measurements")
			// if err != nil {
			// 	App.Logger.Fatalf("error reading event log: %s", err)
			// }
			//App.Logger.Infof("%+v", measurementLog)

			eventLog, err := App.TPM.Measurements(nil)
			if err != nil {
				App.Logger.Fatalf("error parsing event log: %s", err)
			}

			App.Logger.Infof("%+v", eventLog)

			// for _, event := range eventLog.Algs {
			// 	App.Logger.Debug(event)
			// }

			App.Logger.Info("Done...")
		}

	},
}
