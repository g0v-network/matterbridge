// Code generated by msgraph-generate.go DO NOT EDIT.

package msgraph

// WindowsManagementApp Windows management app entity.
type WindowsManagementApp struct {
	// Entity is the base model of WindowsManagementApp
	Entity
	// AvailableVersion Windows management app available version.
	AvailableVersion *string `json:"availableVersion,omitempty"`
	// HealthStates undocumented
	HealthStates []WindowsManagementAppHealthState `json:"healthStates,omitempty"`
}