package main

import (
	"testing"

	terminalv1 "github.com/jraymond/kubernetes-web-terminal/pkg/apis/terminal/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTerminalConfigDeepCopy(t *testing.T) {
	// Create a sample TerminalConfig
	tc := &terminalv1.TerminalConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "terminal.kubernetes-web-terminal.io/v1",
			Kind:       "TerminalConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-terminal",
			Namespace: "default",
		},
		Spec: terminalv1.TerminalConfigSpec{
			Image:   "ubuntu:22.04",
			Command: []string{"/bin/bash"},
			FileMounts: []terminalv1.FileMount{
				{
					Name:      "config-mount",
					MountPath: "/config",
					ConfigMapRef: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "my-config",
						},
					},
					ReadOnly: true,
				},
			},
		},
	}

	// Test DeepCopyObject
	copied := tc.DeepCopyObject()
	copiedTC, ok := copied.(*terminalv1.TerminalConfig)
	if !ok {
		t.Fatalf("DeepCopyObject did not return TerminalConfig type")
	}

	// Verify the copy is not the same object
	if tc == copiedTC {
		t.Errorf("DeepCopyObject returned the same object reference")
	}

	// Verify the content is the same
	if tc.Name != copiedTC.Name {
		t.Errorf("Name mismatch: got %s, want %s", copiedTC.Name, tc.Name)
	}

	if tc.Spec.Image != copiedTC.Spec.Image {
		t.Errorf("Image mismatch: got %s, want %s", copiedTC.Spec.Image, tc.Spec.Image)
	}

	if len(tc.Spec.FileMounts) != len(copiedTC.Spec.FileMounts) {
		t.Errorf("FileMounts length mismatch: got %d, want %d", len(copiedTC.Spec.FileMounts), len(tc.Spec.FileMounts))
	}
}

func TestTerminalConfigListDeepCopy(t *testing.T) {
	// Create a sample TerminalConfigList
	tcList := &terminalv1.TerminalConfigList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "terminal.kubernetes-web-terminal.io/v1",
			Kind:       "TerminalConfigList",
		},
		Items: []terminalv1.TerminalConfig{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1",
				},
				Spec: terminalv1.TerminalConfigSpec{
					Image: "ubuntu:22.04",
				},
			},
		},
	}

	// Test DeepCopyObject
	copied := tcList.DeepCopyObject()
	copiedTCL, ok := copied.(*terminalv1.TerminalConfigList)
	if !ok {
		t.Fatalf("DeepCopyObject did not return TerminalConfigList type")
	}

	// Verify the copy is not the same object
	if tcList == copiedTCL {
		t.Errorf("DeepCopyObject returned the same object reference")
	}

	// Verify the content is the same
	if len(tcList.Items) != len(copiedTCL.Items) {
		t.Errorf("Items length mismatch: got %d, want %d", len(copiedTCL.Items), len(tcList.Items))
	}
}

func TestFileMountTypes(t *testing.T) {
	// Test FileMount with different reference types
	testCases := []struct {
		name      string
		fileMount terminalv1.FileMount
	}{
		{
			name: "ConfigMap reference",
			fileMount: terminalv1.FileMount{
				Name:      "config-mount",
				MountPath: "/config",
				ConfigMapRef: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "my-config",
					},
				},
			},
		},
		{
			name: "Secret reference",
			fileMount: terminalv1.FileMount{
				Name:      "secret-mount",
				MountPath: "/secrets",
				SecretRef: &corev1.SecretVolumeSource{
					SecretName: "my-secret",
				},
			},
		},
		{
			name: "Volume reference",
			fileMount: terminalv1.FileMount{
				Name:      "volume-mount",
				MountPath: "/data",
				VolumeRef: &terminalv1.VolumeReference{
					Name:    "my-volume",
					SubPath: "data",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate that the mount has the expected name and path
			if tc.fileMount.Name == "" {
				t.Errorf("FileMount name should not be empty")
			}
			if tc.fileMount.MountPath == "" {
				t.Errorf("FileMount mountPath should not be empty")
			}

			// Count non-nil references
			refCount := 0
			if tc.fileMount.ConfigMapRef != nil {
				refCount++
			}
			if tc.fileMount.SecretRef != nil {
				refCount++
			}
			if tc.fileMount.VolumeRef != nil {
				refCount++
			}

			if refCount != 1 {
				t.Errorf("FileMount should have exactly one reference type, got %d", refCount)
			}
		})
	}
}