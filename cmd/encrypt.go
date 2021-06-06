package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var EncryptAES bool
var EncryptAESKey string
var EncryptBCrypt bool
var EncryptData string
var EncryptBase64Encode bool

func init() {

	encryptCmd.PersistentFlags().BoolVarP(&EncryptAES, "aes", "", false, "Use AES 256 encryption")
	encryptCmd.PersistentFlags().StringVarP(&EncryptAESKey, "aeskey", "", "", "AES encryption key")
	encryptCmd.PersistentFlags().BoolVarP(&EncryptBCrypt, "bcrypt", "", false, "Use bcrypt encryption")
	encryptCmd.PersistentFlags().BoolVarP(&EncryptBase64Encode, "encode", "", false, "Base64 encode the encrypted data")
	encryptCmd.PersistentFlags().StringVarP(&EncryptData, "data", "", "", "The data to encrypt")

	rootCmd.AddCommand(encryptCmd)
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt data",
	Long:  `Encrypts the specified data using one of the supported algorithms`,
	Run: func(cmd *cobra.Command, args []string) {

		aeskey := []byte(EncryptAESKey)
		if EncryptAES && EncryptAESKey == "" {

			println("generating aes key")

			k := make([]byte, 32)
			_, err := rand.Read(k)
			if err != nil {
				App.Logger.Fatal(err)
			}
			aeskey = k
		}

		/*
			 * RSA encryption
			 *
				keypair, err := app.NewRsaKeyPair(App)
				if err != nil {
					App.Logger.Fatal(err)
				}
		*/

		if EncryptAES {
			crypto := app.NewCrypto(aeskey)
			encrypted, err := crypto.Encrypt([]byte(EncryptData), EncryptBase64Encode)
			if err != nil {
				App.Logger.Fatal(err)
			}
			fmt.Println(string(encrypted))
			os.Exit(0)
		}

		if EncryptBCrypt {
			encrypted, err := bcrypt.GenerateFromPassword([]byte(EncryptData), bcrypt.DefaultCost)
			if err != nil {
				App.Logger.Fatal(err)
			}
			retval := string(encrypted)
			if EncryptBase64Encode {
				retval = base64.StdEncoding.EncodeToString([]byte(retval))
			}
			fmt.Println(retval)
			os.Exit(0)
		}

		App.Logger.Fatal("Invalid command algorithm. Supported algorithms: [ aes | bcrypt ]")
	},
}
