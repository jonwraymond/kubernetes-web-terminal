package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1 "k8s.io/api/core/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TerminalConfig represents a configuration for a terminal session with file mount references
type TerminalConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerminalConfigSpec   `json:"spec,omitempty"`
	Status TerminalConfigStatus `json:"status,omitempty"`
}

// TerminalConfigSpec defines the desired state of TerminalConfig
type TerminalConfigSpec struct {
	// Image specifies the container image to use for the terminal
	// +optional
	Image string `json:"image,omitempty"`

	// Command specifies the command to run in the terminal
	// +optional
	Command []string `json:"command,omitempty"`

	// Args specifies the arguments to pass to the command
	// +optional
	Args []string `json:"args,omitempty"`

	// FileMounts specifies the file mounts to be made available in the terminal
	// +optional
	FileMounts []FileMount `json:"fileMounts,omitempty"`

	// Resources specifies the resource requirements for the terminal container
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// SecurityContext specifies the security context for the terminal container
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
}

// FileMount represents a file mount reference that can be a ConfigMap, Secret, or Volume
type FileMount struct {
	// Name specifies the name of the file mount
	Name string `json:"name"`

	// MountPath specifies where to mount the files in the terminal container
	MountPath string `json:"mountPath"`

	// ConfigMapRef references a ConfigMap to mount
	// +optional
	ConfigMapRef *corev1.ConfigMapVolumeSource `json:"configMapRef,omitempty"`

	// SecretRef references a Secret to mount
	// +optional
	SecretRef *corev1.SecretVolumeSource `json:"secretRef,omitempty"`

	// VolumeRef references an existing Volume to mount
	// +optional
	VolumeRef *VolumeReference `json:"volumeRef,omitempty"`

	// ReadOnly specifies whether the mount should be read-only
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
}

// VolumeReference represents a reference to an existing volume
type VolumeReference struct {
	// Name specifies the name of the volume
	Name string `json:"name"`

	// SubPath specifies a sub-path within the volume
	// +optional
	SubPath string `json:"subPath,omitempty"`
}

// TerminalConfigStatus defines the observed state of TerminalConfig
type TerminalConfigStatus struct {
	// Phase represents the current phase of the terminal configuration
	// +optional
	Phase TerminalConfigPhase `json:"phase,omitempty"`

	// Message provides additional information about the current phase
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions represents the latest available observations of the terminal config's current state
	// +optional
	Conditions []TerminalConfigCondition `json:"conditions,omitempty"`

	// CreatedAt represents when the terminal session was created
	// +optional
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

// TerminalConfigPhase represents the phase of a terminal configuration
type TerminalConfigPhase string

const (
	// TerminalConfigPhasePending indicates the terminal config is pending
	TerminalConfigPhasePending TerminalConfigPhase = "Pending"
	// TerminalConfigPhaseRunning indicates the terminal is running
	TerminalConfigPhaseRunning TerminalConfigPhase = "Running"
	// TerminalConfigPhaseFailed indicates the terminal config failed
	TerminalConfigPhaseFailed TerminalConfigPhase = "Failed"
	// TerminalConfigPhaseTerminated indicates the terminal was terminated
	TerminalConfigPhaseTerminated TerminalConfigPhase = "Terminated"
)

// TerminalConfigCondition describes the state of a terminal config at a certain point
type TerminalConfigCondition struct {
	// Type of terminal config condition
	Type TerminalConfigConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition
	// +optional
	Message string `json:"message,omitempty"`
}

// TerminalConfigConditionType represents the type of condition
type TerminalConfigConditionType string

const (
	// TerminalConfigReady indicates whether the terminal config is ready
	TerminalConfigReady TerminalConfigConditionType = "Ready"
	// TerminalConfigFilesMounted indicates whether the file mounts are ready
	TerminalConfigFilesMounted TerminalConfigConditionType = "FilesMounted"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TerminalConfigList contains a list of TerminalConfig
type TerminalConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TerminalConfig `json:"items"`
}

// DeepCopyObject returns a deep copy of the TerminalConfig for runtime.Object interface
func (tc *TerminalConfig) DeepCopyObject() runtime.Object {
	if tc == nil {
		return nil
	}
	out := new(TerminalConfig)
	tc.deepCopyInto(out)
	return out
}

// DeepCopyObject returns a deep copy of the TerminalConfigList for runtime.Object interface
func (tcl *TerminalConfigList) DeepCopyObject() runtime.Object {
	if tcl == nil {
		return nil
	}
	out := new(TerminalConfigList)
	tcl.deepCopyInto(out)
	return out
}

// deepCopyInto copies all fields from this TerminalConfig into out
func (tc *TerminalConfig) deepCopyInto(out *TerminalConfig) {
	*out = *tc
	out.TypeMeta = tc.TypeMeta
	tc.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	tc.Spec.deepCopyInto(&out.Spec)
	tc.Status.deepCopyInto(&out.Status)
}

// deepCopyInto copies all fields from this TerminalConfigList into out
func (tcl *TerminalConfigList) deepCopyInto(out *TerminalConfigList) {
	*out = *tcl
	out.TypeMeta = tcl.TypeMeta
	tcl.ListMeta.DeepCopyInto(&out.ListMeta)
	if tcl.Items != nil {
		in, out := &tcl.Items, &out.Items
		*out = make([]TerminalConfig, len(*in))
		for i := range *in {
			(*in)[i].deepCopyInto(&(*out)[i])
		}
	}
}

// deepCopyInto copies all fields from this TerminalConfigSpec into out
func (tcs *TerminalConfigSpec) deepCopyInto(out *TerminalConfigSpec) {
	*out = *tcs
	if tcs.Command != nil {
		in, out := &tcs.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if tcs.Args != nil {
		in, out := &tcs.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if tcs.FileMounts != nil {
		in, out := &tcs.FileMounts, &out.FileMounts
		*out = make([]FileMount, len(*in))
		for i := range *in {
			(*in)[i].deepCopyInto(&(*out)[i])
		}
	}
	tcs.Resources.DeepCopyInto(&out.Resources)
	if tcs.SecurityContext != nil {
		in, out := &tcs.SecurityContext, &out.SecurityContext
		*out = new(corev1.SecurityContext)
		(*in).DeepCopyInto(*out)
	}
}

// deepCopyInto copies all fields from this FileMount into out
func (fm *FileMount) deepCopyInto(out *FileMount) {
	*out = *fm
	if fm.ConfigMapRef != nil {
		in, out := &fm.ConfigMapRef, &out.ConfigMapRef
		*out = new(corev1.ConfigMapVolumeSource)
		(*in).DeepCopyInto(*out)
	}
	if fm.SecretRef != nil {
		in, out := &fm.SecretRef, &out.SecretRef
		*out = new(corev1.SecretVolumeSource)
		(*in).DeepCopyInto(*out)
	}
	if fm.VolumeRef != nil {
		in, out := &fm.VolumeRef, &out.VolumeRef
		*out = new(VolumeReference)
		(*in).deepCopyInto(*out)
	}
}

// deepCopyInto copies all fields from this VolumeReference into out
func (vr *VolumeReference) deepCopyInto(out *VolumeReference) {
	*out = *vr
}

// deepCopyInto copies all fields from this TerminalConfigStatus into out
func (tcs *TerminalConfigStatus) deepCopyInto(out *TerminalConfigStatus) {
	*out = *tcs
	if tcs.Conditions != nil {
		in, out := &tcs.Conditions, &out.Conditions
		*out = make([]TerminalConfigCondition, len(*in))
		for i := range *in {
			(*in)[i].deepCopyInto(&(*out)[i])
		}
	}
	if tcs.CreatedAt != nil {
		in, out := &tcs.CreatedAt, &out.CreatedAt
		*out = (*in).DeepCopy()
	}
}

// deepCopyInto copies all fields from this TerminalConfigCondition into out
func (tcc *TerminalConfigCondition) deepCopyInto(out *TerminalConfigCondition) {
	*out = *tcc
	tcc.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}