package auth

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type UserConditionType string
type ConditionStatus string

type User struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec"`
	// +optional
	Status UserStatus `json:"status,omitempty"`
}

// UserSpec defines the desired state of User
type UserSpec struct {
	// Unique email address(https://www.ietf.org/rfc/rfc5322.txt).
	Email string `json:"email"`
	// The preferred written or spoken language for the user.
	// +optional
	Lang string `json:"lang,omitempty"`
	// Description of the user.
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	DisplayName string `json:"displayName,omitempty"`
	// +optional
	Groups []string `json:"groups,omitempty"`
	// password will be encrypted by mutating admission webhook
	EncryptedPassword string `json:"password,omitempty"`
}

type UserState string

// UserStatus defines the observed state of User
type UserStatus struct {
	// The user status
	// +optional
	State UserState `json:"state,omitempty"`
	// Represents the latest available observations of a user's current state.
	// +optional
	Conditions []UserCondition `json:"conditions,omitempty"`
}

type UserCondition struct {
	// Type of user controller condition.
	Type UserConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

type UserMap struct {
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			SelfLink          string    `json:"selfLink"`
			UID               string    `json:"uid"`
			ResourceVersion   string    `json:"resourceVersion"`
			Generation        int       `json:"generation"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Labels            struct {
				KubefedIoManaged string `json:"kubefed.io/managed"`
			} `json:"labels"`
			Annotations struct {
				IamKubesphereIoGlobalrole        string `json:"iam.kubesphere.io/globalrole"`
				IamKubesphereIoPasswordEncrypted string `json:"iam.kubesphere.io/password-encrypted"`
			} `json:"annotations"`
			Finalizers []string `json:"finalizers"`
		} `json:"metadata"`
		Spec struct {
			Email string `json:"email"`
			Lang  string `json:"lang"`
		} `json:"spec"`
		Status struct {
			State              string    `json:"state"`
			LastTransitionTime time.Time `json:"lastTransitionTime"`
			LastLoginTime      time.Time `json:"lastLoginTime"`
		} `json:"status"`
	} `json:"items"`
	TotalItems int `json:"totalItems"`
}
