package main

import (
	"encoding/json"
	"testing"

	terminalv1 "github.com/jraymond/kubernetes-web-terminal/pkg/apis/terminal/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mustParseQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		panic(err)
	}
	return q
}

func TestTerminalConfigJSONSerialization(t *testing.T) {
	// Create a TerminalConfig with all FileMount types
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
			Args:    []string{"-c", "echo 'Hello World'"},
			FileMounts: []terminalv1.FileMount{
				{
					Name:      "config-mount",
					MountPath: "/etc/config",
					ConfigMapRef: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "app-config",
						},
						Optional: &[]bool{false}[0],
					},
					ReadOnly: true,
				},
				{
					Name:      "secret-mount",
					MountPath: "/etc/secrets",
					SecretRef: &corev1.SecretVolumeSource{
						SecretName: "app-secrets",
						Optional:   &[]bool{false}[0],
					},
					ReadOnly: true,
				},
				{
					Name:      "volume-mount",
					MountPath: "/data",
					VolumeRef: &terminalv1.VolumeReference{
						Name:    "shared-data",
						SubPath: "app-data",
					},
					ReadOnly: false,
				},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: mustParseQuantity("128Mi"),
					corev1.ResourceCPU:    mustParseQuantity("100m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: mustParseQuantity("256Mi"),
					corev1.ResourceCPU:    mustParseQuantity("200m"),
				},
			},
		},
		Status: terminalv1.TerminalConfigStatus{
			Phase:   terminalv1.TerminalConfigPhasePending,
			Message: "Terminal configuration created",
			Conditions: []terminalv1.TerminalConfigCondition{
				{
					Type:   terminalv1.TerminalConfigReady,
					Status: corev1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason: "Pending",
					Message: "Waiting for terminal to be ready",
				},
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.MarshalIndent(tc, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal TerminalConfig to JSON: %v", err)
	}

	t.Logf("TerminalConfig JSON:\n%s", string(jsonData))

	// Test JSON unmarshaling
	var unmarshaledTC terminalv1.TerminalConfig
	err = json.Unmarshal(jsonData, &unmarshaledTC)
	if err != nil {
		t.Fatalf("Failed to unmarshal TerminalConfig from JSON: %v", err)
	}

	// Verify key fields
	if unmarshaledTC.Name != tc.Name {
		t.Errorf("Name mismatch after JSON round-trip: got %s, want %s", unmarshaledTC.Name, tc.Name)
	}

	if unmarshaledTC.Spec.Image != tc.Spec.Image {
		t.Errorf("Image mismatch after JSON round-trip: got %s, want %s", unmarshaledTC.Spec.Image, tc.Spec.Image)
	}

	if len(unmarshaledTC.Spec.FileMounts) != len(tc.Spec.FileMounts) {
		t.Errorf("FileMounts length mismatch after JSON round-trip: got %d, want %d", 
			len(unmarshaledTC.Spec.FileMounts), len(tc.Spec.FileMounts))
	}

	// Verify specific file mount details
	for i, mount := range unmarshaledTC.Spec.FileMounts {
		originalMount := tc.Spec.FileMounts[i]
		
		if mount.Name != originalMount.Name {
			t.Errorf("FileMount[%d] name mismatch: got %s, want %s", i, mount.Name, originalMount.Name)
		}
		
		if mount.MountPath != originalMount.MountPath {
			t.Errorf("FileMount[%d] mountPath mismatch: got %s, want %s", i, mount.MountPath, originalMount.MountPath)
		}
		
		if mount.ReadOnly != originalMount.ReadOnly {
			t.Errorf("FileMount[%d] readOnly mismatch: got %v, want %v", i, mount.ReadOnly, originalMount.ReadOnly)
		}
	}

	// Test that at least one of each mount type exists
	hasConfigMap := false
	hasSecret := false
	hasVolume := false
	
	for _, mount := range unmarshaledTC.Spec.FileMounts {
		if mount.ConfigMapRef != nil {
			hasConfigMap = true
		}
		if mount.SecretRef != nil {
			hasSecret = true
		}
		if mount.VolumeRef != nil {
			hasVolume = true
		}
	}

	if !hasConfigMap {
		t.Error("Expected at least one ConfigMap file mount")
	}
	if !hasSecret {
		t.Error("Expected at least one Secret file mount")
	}
	if !hasVolume {
		t.Error("Expected at least one Volume file mount")
	}
}

func TestTerminalConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		tc          *terminalv1.TerminalConfig
		expectValid bool
	}{
		{
			name: "Valid basic config",
			tc: &terminalv1.TerminalConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "valid-terminal",
				},
				Spec: terminalv1.TerminalConfigSpec{
					Image:   "ubuntu:22.04",
					Command: []string{"/bin/bash"},
				},
			},
			expectValid: true,
		},
		{
			name: "Valid config with file mounts",
			tc: &terminalv1.TerminalConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "terminal-with-mounts",
				},
				Spec: terminalv1.TerminalConfigSpec{
					Image:   "alpine:latest",
					Command: []string{"/bin/sh"},
					FileMounts: []terminalv1.FileMount{
						{
							Name:      "config",
							MountPath: "/config",
							ConfigMapRef: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
							},
						},
					},
				},
			},
			expectValid: true,
		},
		{
			name: "Config with empty name",
			tc: &terminalv1.TerminalConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "",
				},
				Spec: terminalv1.TerminalConfigSpec{
					Image: "ubuntu:22.04",
				},
			},
			expectValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Basic validation - check if required fields are present
			isValid := tc.tc.Name != ""
			
			if isValid != tc.expectValid {
				t.Errorf("Expected validation result %v, got %v for %s", tc.expectValid, isValid, tc.name)
			}
		})
	}
}