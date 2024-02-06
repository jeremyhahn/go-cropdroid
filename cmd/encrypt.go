package cmd

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/jeremyhahn/go-cropdroid/app"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/spf13/cobra"
)

var EncryptAES bool
var EncryptAESKey string
var EncryptArgon2 bool
var EncryptData string
var EncryptBase64Encode bool

func init() {

	encryptCmd.PersistentFlags().BoolVarP(&EncryptAES, "aes", "", false, "Use AES 256 encryption")
	encryptCmd.PersistentFlags().StringVarP(&EncryptAESKey, "aeskey", "", "", "AES encryption key")
	encryptCmd.PersistentFlags().BoolVarP(&EncryptArgon2, "argon2", "", false, "Use argon2 encryption")
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

		if EncryptArgon2 {
			hasher := util.CreatePasswordHasher(App.PasswordHasherParams)
			hash, err := hasher.Encrypt(EncryptData)
			if err != nil {
				App.Logger.Fatal(err)
			}
			fmt.Println(hash)
			os.Exit(0)
		}

		App.Logger.Fatal("Invalid crypto operation. Supported algorithms: [ aes | argon2 ]")
	},
}
