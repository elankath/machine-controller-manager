// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package nodeops is used to provide the node functionalities
package nodeops

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	taintutils "github.com/gardener/machine-controller-manager/pkg/util/taints"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	clientretry "k8s.io/client-go/util/retry"
)

// Backoff is the backoff period used while updating nodes
var Backoff = wait.Backoff{
	Steps:    5,
	Duration: 100 * time.Millisecond,
	Jitter:   1.0,
}

// AddOrUpdateTaintOnNode add taints to the node. If taint was added into node, it'll issue API calls
// to update nodes; otherwise, no API calls. Return error if any.
func AddOrUpdateTaintOnNode(ctx context.Context, c clientset.Interface, nodeName string, taints ...*v1.Taint) error {
	if len(taints) == 0 {
		return nil
	}
	firstTry := true
	return clientretry.RetryOnConflict(Backoff, func() error {
		var err error
		var oldNode *v1.Node
		// First we try getting node from the API server cache, as it's cheaper. If it fails
		// we get it from etcd to be sure to have fresh data.
		if firstTry {
			oldNode, err = c.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{ResourceVersion: "0"})
			firstTry = false
		} else {
			oldNode, err = c.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		}
		if err != nil {
			return err
		}

		var newNode *v1.Node
		oldNodeCopy := oldNode
		updated := false
		for _, taint := range taints {
			curNewNode, ok, err := taintutils.AddOrUpdateTaint(oldNodeCopy, taint)
			if err != nil {
				return fmt.Errorf("failed to update taint of node")
			}
			updated = updated || ok
			newNode = curNewNode
			oldNodeCopy = curNewNode
		}
		if !updated {
			return nil
		}
		return UpdateNodeTaints(ctx, c, nodeName, oldNode, newNode)
	})
}

// RemoveTaintOffNode is for cleaning up taints temporarily added to node,
// won't fail if target taint doesn't exist or has been removed.
// If passed a node it'll check if there's anything to be done, if taint is not present it won't issue
// any API calls.
func RemoveTaintOffNode(ctx context.Context, c clientset.Interface, nodeName string, node *v1.Node, taints ...*v1.Taint) error {
	if len(taints) == 0 {
		return nil
	}
	// Short circuit for limiting amount of API calls.
	if node != nil {
		match := false
		for _, taint := range taints {
			if taintutils.TaintExists(node.Spec.Taints, taint) {
				match = true
				break
			}
		}
		if !match {
			return nil
		}
	}

	firstTry := true
	return clientretry.RetryOnConflict(Backoff, func() error {
		var err error
		var oldNode *v1.Node
		// First we try getting node from the API server cache, as it's cheaper. If it fails
		// we get it from etcd to be sure to have fresh data.
		if firstTry {
			oldNode, err = c.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{ResourceVersion: "0"})
			firstTry = false
		} else {
			oldNode, err = c.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		}
		if err != nil {
			return err
		}

		var newNode *v1.Node
		oldNodeCopy := oldNode
		updated := false
		for _, taint := range taints {
			curNewNode, ok, err := taintutils.RemoveTaint(oldNodeCopy, taint)
			if err != nil {
				return fmt.Errorf("failed to remove taint of node")
			}
			updated = updated || ok
			newNode = curNewNode
			oldNodeCopy = curNewNode
		}
		if !updated {
			return nil
		}
		return UpdateNodeTaints(ctx, c, nodeName, oldNode, newNode)
	})
}

// PatchNodeTaints is for updating the node taints from oldNode to the newNode
// It makes a TwoWayMergePatch by comparing the two objects
// It calls the Patch() method to do the final patch
func PatchNodeTaints(ctx context.Context, c clientset.Interface, nodeName string, oldNode *v1.Node, newNode *v1.Node) error {
	oldData, err := json.Marshal(oldNode)
	if err != nil {
		return fmt.Errorf("failed to marshal old node %#v for node %q: %v", oldNode, nodeName, err)
	}

	newTaints := newNode.Spec.Taints
	newNodeClone := oldNode.DeepCopy()
	newNodeClone.Spec.Taints = newTaints
	newData, err := json.Marshal(newNodeClone)
	if err != nil {
		return fmt.Errorf("failed to marshal new node %#v for node %q: %v", newNodeClone, nodeName, err)
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Node{})
	if err != nil {
		return fmt.Errorf("failed to create patch for node %q: %v", nodeName, err)
	}

	_, err = c.CoreV1().Nodes().Patch(ctx, string(nodeName), types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	return err
}

// UpdateNodeTaints is for updating the node taints from oldNode to the newNode
// using the nodes Update() method
func UpdateNodeTaints(ctx context.Context, c clientset.Interface, nodeName string, oldNode *v1.Node, newNode *v1.Node) error {
	newNodeClone := oldNode.DeepCopy()
	newNodeClone.Spec.Taints = newNode.Spec.Taints

	_, err := c.CoreV1().Nodes().Update(ctx, newNodeClone, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create update taints for node %q: %v", nodeName, err)
	}

	return nil
}
