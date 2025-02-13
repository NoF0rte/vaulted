package menu

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/fatih/color"

	vaulted "github.com/NoF0rte/vaulted/v3/lib"
)

// AWSMenu The menu type for the AWS edit tree
type AWSMenu struct {
	*Menu
}

func sdkRegions() map[string]bool {
	regions := make(map[string]bool)
	for _, partition := range endpoints.DefaultPartitions() {
		for region := range partition.Regions() {
			regions[region] = true
		}
	}
	return regions
}

func (m *AWSMenu) Help() {
	menuColor.Set()
	defer color.Unset()

	fmt.Println("k,key    - Key")
	fmt.Println("m,mfa    - MFA")
	fmt.Println("r,role   - Role")
	fmt.Println("R,region - Region")
	fmt.Println("t,temp   - Substitute with temporary credentials")
	fmt.Println("S,show   - Show/Hide Secrets")
	fmt.Println("D,delete - Delete")
	fmt.Println("?,help   - Help")
	fmt.Println("b,back   - Back")
	fmt.Println("q,quit   - Quit")
}

func (m *AWSMenu) Handler() error {
	var err error

	for {
		var input string
		m.Printer()
		if m.Vault.AWSKey == nil {
			input, err = interaction.ReadMenu("Edit AWS key [k,b]: ")
		} else {
			input, err = interaction.ReadMenu("Edit AWS key [k,m,r,R,t,S,D,b]: ")
		}

		if err != nil {
			return err
		}

		switch input {
		case "k", "add", "key", "keys":
			warningColor.Println("Note: For increased security, Vaulted defaults to substituting your credentials with temporary credentials.")
			warningColor.Println("      The key specified here may not match the key in your spawned session.")
			fmt.Println("")

			awsAccesskey, keyErr := interaction.ReadValue("Key ID: ")
			if keyErr != nil {
				return keyErr
			}
			awsSecretkey, secretErr := interaction.ReadValue("Secret: ")
			if secretErr != nil {
				return secretErr
			}

			// Copy the key to prevent other values from being lost
			oldAWSKey := m.Vault.AWSKey
			if oldAWSKey == nil {
				m.Vault.AWSKey = &vaulted.AWSKey{}
			} else {
				newAWSKey := *oldAWSKey
				m.Vault.AWSKey = &newAWSKey
			}

			m.Vault.AWSKey.ID = awsAccesskey
			m.Vault.AWSKey.Secret = awsSecretkey

			// Attempt to autoconfigure the MFA
			detectMFAMenu := DetectMFAMenu{Menu: m.Menu}
			detectErr := detectMFAMenu.Handler()
			if detectErr == ErrUserAbort {
				m.Vault.AWSKey = oldAWSKey
			}
		case "m", "mfa":
			if m.Vault.AWSKey != nil {
				// Save the old MFA in case the user aborts
				oldMFA := m.Vault.AWSKey.MFA
				m.Vault.AWSKey.MFA = ""

				detectMFAMenu := DetectMFAMenu{Menu: m.Menu}
				detectErr := detectMFAMenu.Handler()
				if detectErr == ErrUserAbort {
					m.Vault.AWSKey.MFA = oldMFA
					continue
				}

				if m.Vault.AWSKey.MFA == "" {
					var awsMfa string
					awsMfa, err = interaction.ReadValue("MFA ARN or serial number: ")
					if err == ErrUserAbort {
						m.Vault.AWSKey.MFA = oldMFA
						continue
					} else if err == nil {
						m.Vault.AWSKey.MFA = awsMfa
					}
				}
			} else {
				color.Red("Must associate an AWS key with the vault first")
			}
		case "r", "role":
			if m.Vault.AWSKey != nil {
				var awsRole string
				awsRole, err = interaction.ReadValue("Role ARN: ")
				if err == nil {
					m.Vault.AWSKey.Role = awsRole
				}
			} else {
				color.Red("Must associate an AWS key with the vault first")
			}
		case "R", "region":
			region, err := m.readRegion()
			if err != nil {
				return err
			}

			if region != "" {
				// Set the region
				if m.Vault.AWSKey == nil {
					m.Vault.AWSKey = &vaulted.AWSKey{}
				}
				m.Vault.AWSKey.Region = &region
			} else {
				// Unset the region
				if m.Vault.AWSKey != nil {
					m.Vault.AWSKey.Region = nil
				}
			}

		case "t", "temp", "temporary":
			if m.Vault.AWSKey != nil {
				forgoTempCredGeneration := !m.Vault.AWSKey.ForgoTempCredGeneration
				if !forgoTempCredGeneration && m.Vault.Duration > 36*time.Hour {
					var conf string
					warningColor.Println("Proceeding will adjust your vault duration to 36h (the maximum when using temporary creds).")
					conf, err = interaction.ReadPrompt("Do you wish to proceed? (y/n): ")
					if conf == "y" {
						m.Vault.Duration = 36 * time.Hour
					} else {
						fmt.Println("Temporary credentials not enabled.")
						continue
					}
				}

				m.Vault.AWSKey.ForgoTempCredGeneration = forgoTempCredGeneration
			} else {
				color.Red("Must associate an AWS key with the vault first")
			}
		case "S", "show", "hide":
			m.toggleHidden()
		case "D", "delete", "remove":
			if m.Vault.AWSKey != nil {
				var removeKey string
				removeKey, err = interaction.ReadValue("Delete your AWS key? (y/n): ")
				if err == nil {
					if removeKey == "y" {
						m.Vault.AWSKey = nil
					}
				}
			} else {
				color.Red("Must associate an AWS key with the vault first")
			}
		case "b", "back":
			return nil
		case "q", "quit", "exit":
			var confirm string
			confirm, err = interaction.ReadValue("Are you sure you wish to save and exit the vault? (y/n): ")
			if err == nil {
				if confirm == "y" {
					return ErrSaveAndExit
				}
			}
		case "?", "help":
			m.Help()
		default:
			color.Red("Command not recognized")
		}

		if err != nil {
			return err
		}
	}
}

func (m *AWSMenu) readRegion() (string, error) {
	region, err := interaction.ReadValue("Region: ")
	if err != nil {
		return "", err
	}

	if !sdkRegions()[region] {
		fmt.Printf("\n%s%s%s\n", warningColor.Sprint("WARNING: "), region, warningColor.Sprint(" doesn't appear to be a valid region."))
	}

	return region, nil
}

func (m *AWSMenu) Printer() {
	color.Cyan("\nAWS Key:")
	if m.Vault.AWSKey != nil {
		green.Printf("  Key ID: ")
		fmt.Printf("%s\n", m.Vault.AWSKey.ID)
		green.Printf("  Secret: ")
		if m.Menu.ShowHidden {
			fmt.Printf("%s\n", m.Vault.AWSKey.Secret)
		} else {
			fmt.Printf("%s\n", faintColor.Sprint("<hidden>"))
		}
		green.Printf("  MFA: ")
		if m.Vault.AWSKey.MFA == "" {
			var warning string
			if !m.Vault.AWSKey.ForgoTempCredGeneration {
				warning = warningColor.Sprint(" (warning: some APIs will not function without MFA (e.g. IAM))")
			}
			fmt.Printf("%s %s\n", faintColor.Sprint("<not configured>"), warning)
		} else {
			fmt.Printf("%s\n", m.Vault.AWSKey.MFA)
		}
		if m.Vault.AWSKey.Role != "" {
			green.Printf("  Role: ")
			fmt.Printf("%s\n", m.Vault.AWSKey.Role)
		}
		green.Printf("  Substitute with temporary credentials: ")
		fmt.Printf("%t\n", !m.Vault.AWSKey.ForgoTempCredGeneration)

		green.Printf("  Region: ")
		if m.Vault.AWSKey.Region == nil || *m.Vault.AWSKey.Region == "" {
			fmt.Printf("%s\n", faintColor.Sprint("<global>"))
		} else {
			var unrecognized string
			if !sdkRegions()[*m.Vault.AWSKey.Region] {
				unrecognized = warningColor.Sprintf(" (unrecognized region)")
			}

			fmt.Printf("%s%s\n", *m.Vault.AWSKey.Region, unrecognized)
		}
	} else {
		fmt.Println("  [Empty]")
	}
}
