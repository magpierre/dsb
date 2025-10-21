// Copyright 2025 Magnus Pierre
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package windows

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

// TreeNodeType represents the type of node in the navigation tree
type TreeNodeType string

const (
	NodeTypeShare  TreeNodeType = "share"
	NodeTypeSchema TreeNodeType = "schema"
	NodeTypeTable  TreeNodeType = "table"
)

// TreeNode represents a node in the navigation tree
type TreeNode struct {
	ID           string              // Unique identifier
	NodeType     TreeNodeType        // Type of node
	Name         string              // Display name
	Share        string              // Parent share name
	Schema       string              // Parent schema name (for tables)
	Table        delta_sharing.Table // Full table object (for table nodes)
	Children     []string            // Child node IDs
	ChildrenLoaded bool              // Whether children have been loaded from server
}

// NavigationTree manages the hierarchical tree structure for Delta Sharing navigation
type NavigationTree struct {
	nodes       map[string]*TreeNode
	rootIDs     []string
	profile     string
	client      delta_sharing.SharingClient
	mainWin     *MainWindow
	mu          sync.RWMutex // Protect concurrent access during lazy loading
}

// NewNavigationTree creates and initializes a new navigation tree
func NewNavigationTree(mainWin *MainWindow) *NavigationTree {
	return &NavigationTree{
		nodes:   make(map[string]*TreeNode),
		rootIDs: make([]string, 0),
		mainWin: mainWin,
	}
}

// GenerateNodeID creates a unique ID for a tree node
func (nt *NavigationTree) GenerateNodeID(nodeType TreeNodeType, share, schema, table string) string {
	switch nodeType {
	case NodeTypeShare:
		return fmt.Sprintf("share:%s", share)
	case NodeTypeSchema:
		return fmt.Sprintf("share:%s:schema:%s", share, schema)
	case NodeTypeTable:
		return fmt.Sprintf("share:%s:schema:%s:table:%s", share, schema, table)
	default:
		return ""
	}
}

// ParseNodeID extracts components from a node ID
func (nt *NavigationTree) ParseNodeID(nodeID string) (nodeType TreeNodeType, share, schema, table string) {
	parts := strings.Split(nodeID, ":")

	if len(parts) >= 2 && parts[0] == "share" {
		nodeType = NodeTypeShare
		share = parts[1]
	}

	if len(parts) >= 4 && parts[2] == "schema" {
		nodeType = NodeTypeSchema
		schema = parts[3]
	}

	if len(parts) >= 6 && parts[4] == "table" {
		nodeType = NodeTypeTable
		table = parts[5]
	}

	return
}

// LoadShares populates the tree with root-level share nodes and preloads all tables
func (nt *NavigationTree) LoadShares(profile string) error {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	nt.profile = profile

	// Create Delta Sharing client
	client, err := delta_sharing.NewSharingClientFromString(profile)
	if err != nil {
		return fmt.Errorf("failed to create Delta Sharing client: %w", err)
	}
	nt.client = client

	// Fetch shares from server
	shares, _, err := client.ListShares(context.Background(), 0, "")
	if err != nil {
		return fmt.Errorf("failed to list shares: %w", err)
	}

	// Clear existing tree
	nt.nodes = make(map[string]*TreeNode)
	nt.rootIDs = make([]string, 0, len(shares))

	// Create share nodes
	shareMap := make(map[string]*TreeNode)
	for _, share := range shares {
		nodeID := nt.GenerateNodeID(NodeTypeShare, share.Name, "", "")
		node := &TreeNode{
			ID:             nodeID,
			NodeType:       NodeTypeShare,
			Name:           share.Name,
			Share:          share.Name,
			Children:       make([]string, 0),
			ChildrenLoaded: true, // Will be populated below
		}
		nt.nodes[nodeID] = node
		nt.rootIDs = append(nt.rootIDs, nodeID)
		shareMap[share.Name] = node
	}

	// Preload all tables using ListAllTables
	allTables, _, err := client.ListAllTables(context.Background(), 0, "")
	if err != nil {
		return fmt.Errorf("failed to list all tables: %w", err)
	}

	// Map to track schema nodes by their ID
	schemaMap := make(map[string]*TreeNode)

	// Create schema and table nodes from the preloaded data
	for _, table := range allTables {
		shareName := table.Share
		schemaName := table.Schema
		tableName := table.Name

		// Get or create share node (should already exist)
		shareNode, shareExists := shareMap[shareName]
		if !shareExists {
			// If share doesn't exist, create it
			shareNodeID := nt.GenerateNodeID(NodeTypeShare, shareName, "", "")
			shareNode = &TreeNode{
				ID:             shareNodeID,
				NodeType:       NodeTypeShare,
				Name:           shareName,
				Share:          shareName,
				Children:       make([]string, 0),
				ChildrenLoaded: true,
			}
			nt.nodes[shareNodeID] = shareNode
			nt.rootIDs = append(nt.rootIDs, shareNodeID)
			shareMap[shareName] = shareNode
		}

		// Get or create schema node
		schemaNodeID := nt.GenerateNodeID(NodeTypeSchema, shareName, schemaName, "")
		schemaNode, schemaExists := schemaMap[schemaNodeID]
		if !schemaExists {
			schemaNode = &TreeNode{
				ID:             schemaNodeID,
				NodeType:       NodeTypeSchema,
				Name:           schemaName,
				Share:          shareName,
				Schema:         schemaName,
				Children:       make([]string, 0),
				ChildrenLoaded: true,
			}
			nt.nodes[schemaNodeID] = schemaNode
			schemaMap[schemaNodeID] = schemaNode
			// Add schema to share's children
			shareNode.Children = append(shareNode.Children, schemaNodeID)
		}

		// Create table node
		tableNodeID := nt.GenerateNodeID(NodeTypeTable, shareName, schemaName, tableName)
		tableNode := &TreeNode{
			ID:             tableNodeID,
			NodeType:       NodeTypeTable,
			Name:           tableName,
			Share:          shareName,
			Schema:         schemaName,
			Table:          table,
			Children:       nil,
			ChildrenLoaded: true, // Tables don't have children
		}
		nt.nodes[tableNodeID] = tableNode
		// Add table to schema's children
		schemaNode.Children = append(schemaNode.Children, tableNodeID)
	}

	return nil
}

// GetChildren returns the child node IDs for a given parent node
// Returns root nodes if nodeID is empty
func (nt *NavigationTree) GetChildren(nodeID widget.TreeNodeID) []widget.TreeNodeID {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	// Root level - return shares
	if nodeID == "" {
		return nt.rootIDs
	}

	// Get node and return its children
	node, exists := nt.nodes[nodeID]
	if !exists {
		return []widget.TreeNodeID{}
	}

	return node.Children
}

// IsBranch returns true if the node can have children
func (nt *NavigationTree) IsBranch(nodeID widget.TreeNodeID) bool {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	// Root is always a branch
	if nodeID == "" {
		return true
	}

	node, exists := nt.nodes[nodeID]
	if !exists {
		return false
	}

	// Shares and schemas are branches, tables are leaves
	return node.NodeType == NodeTypeShare || node.NodeType == NodeTypeSchema
}

// GetNode retrieves a node by ID
func (nt *NavigationTree) GetNode(nodeID widget.TreeNodeID) *TreeNode {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	return nt.nodes[nodeID]
}

// LazyLoadChildren is no longer needed since all data is preloaded
// This method is kept for compatibility but will always return early
func (nt *NavigationTree) LazyLoadChildren(nodeID widget.TreeNodeID) error {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	node, exists := nt.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// All children are preloaded, so this should always be true
	if node.ChildrenLoaded {
		return nil
	}

	// This shouldn't happen with preloading, but handle gracefully
	return nil
}

// loadSchemas fetches schemas for a share node
func (nt *NavigationTree) loadSchemas(shareNode *TreeNode) error {
	// Find the share object
	shares, _, err := nt.client.ListShares(context.Background(), 0, "")
	if err != nil {
		return fmt.Errorf("failed to list shares: %w", err)
	}

	// Use type inference - share type is unexported
	var shareObj = shares[0] // Will be replaced by actual match
	found := false
	for _, s := range shares {
		if s.Name == shareNode.Share {
			shareObj = s
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("share not found: %s", shareNode.Share)
	}

	// Fetch schemas
	schemas, _, err := nt.client.ListSchemas(context.Background(), shareObj, 0, "")
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}

	// Create schema nodes
	shareNode.Children = make([]string, 0, len(schemas))
	for _, schema := range schemas {
		nodeID := nt.GenerateNodeID(NodeTypeSchema, shareNode.Share, schema.Name, "")
		schemaNode := &TreeNode{
			ID:             nodeID,
			NodeType:       NodeTypeSchema,
			Name:           schema.Name,
			Share:          shareNode.Share,
			Schema:         schema.Name,
			Children:       nil, // Will be lazy-loaded
			ChildrenLoaded: false,
		}
		nt.nodes[nodeID] = schemaNode
		shareNode.Children = append(shareNode.Children, nodeID)
	}

	shareNode.ChildrenLoaded = true
	return nil
}

// loadTables fetches tables for a schema node
func (nt *NavigationTree) loadTables(schemaNode *TreeNode) error {
	// Find the share and schema objects
	shares, _, err := nt.client.ListShares(context.Background(), 0, "")
	if err != nil {
		return fmt.Errorf("failed to list shares: %w", err)
	}

	// Use type inference - share type is unexported
	var shareObj = shares[0] // Will be replaced by actual match
	for _, s := range shares {
		if s.Name == schemaNode.Share {
			shareObj = s
			break
		}
	}

	schemas, _, err := nt.client.ListSchemas(context.Background(), shareObj, 0, "")
	if err != nil {
		return fmt.Errorf("failed to list schemas: %w", err)
	}

	// Use type inference - schema type is unexported
	var schemaObj = schemas[0] // Will be replaced by actual match
	found := false
	for _, s := range schemas {
		if s.Name == schemaNode.Schema {
			schemaObj = s
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("schema not found: %s", schemaNode.Schema)
	}

	// Fetch tables
	tables, _, err := nt.client.ListTables(context.Background(), schemaObj, 0, "")
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	// Create table nodes
	schemaNode.Children = make([]string, 0, len(tables))
	for _, table := range tables {
		nodeID := nt.GenerateNodeID(NodeTypeTable, schemaNode.Share, schemaNode.Schema, table.Name)
		tableNode := &TreeNode{
			ID:             nodeID,
			NodeType:       NodeTypeTable,
			Name:           table.Name,
			Share:          schemaNode.Share,
			Schema:         schemaNode.Schema,
			Table:          table,
			Children:       nil,
			ChildrenLoaded: true, // Tables don't have children
		}
		nt.nodes[nodeID] = tableNode
		schemaNode.Children = append(schemaNode.Children, nodeID)
	}

	schemaNode.ChildrenLoaded = true
	return nil
}

// UpdateNodeDisplay updates the visual representation of a tree node
func (nt *NavigationTree) UpdateNodeDisplay(nodeID widget.TreeNodeID, obj fyne.CanvasObject, branch bool) {
	node := nt.GetNode(nodeID)
	if node == nil {
		return
	}

	// Get the container and its children
	box, ok := obj.(*fyne.Container)
	if !ok || len(box.Objects) < 2 {
		return
	}

	// Update icon
	icon, ok := box.Objects[0].(*widget.Icon)
	if ok {
		switch node.NodeType {
		case NodeTypeShare:
			icon.SetResource(theme.FolderOpenIcon())
		case NodeTypeSchema:
			icon.SetResource(theme.FolderIcon())
		case NodeTypeTable:
			icon.SetResource(theme.DocumentIcon())
		}
	}

	// Update label
	label, ok := box.Objects[1].(*widget.Label)
	if ok {
		label.SetText(node.Name)
	}
}
