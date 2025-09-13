package score

import (
	"context"
	"fmt"
	"math"
)

// NetworkGraphAnalyzer implements GraphAnalyzer using network analysis algorithms
type NetworkGraphAnalyzer struct {
	dataProvider DataProvider
}

// NewNetworkGraphAnalyzer creates a new network graph analyzer
func NewNetworkGraphAnalyzer(dataProvider DataProvider) GraphAnalyzer {
	return &NetworkGraphAnalyzer{
		dataProvider: dataProvider,
	}
}

// VouchEdge represents an edge in the vouch graph
type VouchEdge struct {
	From     string
	To       string
	Weight   float64
	Strength float64
	Epoch    int64
}

// VouchGraph represents the vouch graph for analysis
type VouchGraph struct {
	Nodes map[string]bool         // Set of all nodes (DIDs)
	Edges []VouchEdge             // All vouch edges
	AdjList map[string][]VouchEdge // Adjacency list representation
}

// buildVouchGraph constructs the vouch graph from vouches data
func (g *NetworkGraphAnalyzer) buildVouchGraph(ctx context.Context, context string, epoch int64) (*VouchGraph, error) {
	graph := &VouchGraph{
		Nodes:   make(map[string]bool),
		Edges:   make([]VouchEdge, 0),
		AdjList: make(map[string][]VouchEdge),
	}
	
	// This is a simplified approach - in practice, we'd need to get all vouches
	// For now, we'll build incrementally as we encounter DIDs
	// This would typically be pre-computed and cached
	
	return graph, nil
}

// DetectCollusion implements GraphAnalyzer.DetectCollusion
func (g *NetworkGraphAnalyzer) DetectCollusion(ctx context.Context, context string, epoch int64) ([]*CollusionCluster, error) {
	graph, err := g.buildVouchGraph(ctx, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to build vouch graph: %w", err)
	}
	
	// Find dense subgraphs that might indicate collusion
	denseSubgraphs, err := g.findDenseSubgraphs(graph, 0.7) // 70% density threshold
	if err != nil {
		return nil, fmt.Errorf("failed to find dense subgraphs: %w", err)
	}
	
	// Convert dense subgraphs to collusion clusters
	clusters := make([]*CollusionCluster, 0, len(denseSubgraphs))
	for _, subgraph := range denseSubgraphs {
		if len(subgraph.Nodes) >= 3 { // Require at least 3 nodes for collusion
			cluster := &CollusionCluster{
				Context:     context,
				Epoch:       epoch,
				Members:     subgraph.Nodes,
				Density:     subgraph.Density,
				VouchVolume: g.calculateVouchVolume(subgraph.Nodes, graph),
				Confidence:  g.calculateCollusionConfidence(subgraph),
			}
			clusters = append(clusters, cluster)
		}
	}
	
	return clusters, nil
}

// ComputeDiversity implements GraphAnalyzer.ComputeDiversity
func (g *NetworkGraphAnalyzer) ComputeDiversity(ctx context.Context, did, context string, epoch int64) (float64, error) {
	// Get vouches for this DID
	vouches, err := g.dataProvider.GetVouches(ctx, did, context, epoch)
	if err != nil {
		return 1.0, err // Default to maximum diversity on error
	}
	
	if len(vouches) == 0 {
		return 1.0, nil // Maximum diversity if no vouches
	}
	
	// Compute diversity using Shannon entropy or Gini-Simpson index
	return g.computeShannnonDiversity(ctx, did, vouches, context, epoch)
}

// computeShannnonDiversity computes diversity using Shannon entropy
func (g *NetworkGraphAnalyzer) computeShannnonDiversity(ctx context.Context, targetDID string, vouches []*VouchData, context string, epoch int64) (float64, error) {
	// Group vouchers by community/cluster
	communities, err := g.identifyCommunities(ctx, vouches, context, epoch)
	if err != nil {
		return 1.0, nil // Default to high diversity on error
	}
	
	// Count vouches per community
	communityVouches := make(map[string]int)
	totalVouches := len(vouches)
	
	for _, vouch := range vouches {
		community := communities[vouch.FromDID]
		if community == "" {
			community = "unknown"
		}
		communityVouches[community]++
	}
	
	// Calculate Shannon entropy
	entropy := 0.0
	for _, count := range communityVouches {
		if count > 0 {
			p := float64(count) / float64(totalVouches)
			entropy -= p * math.Log2(p)
		}
	}
	
	// Normalize entropy to [0, 1]
	maxEntropy := math.Log2(float64(len(communityVouches)))
	if maxEntropy == 0 {
		return 1.0, nil
	}
	
	return entropy / maxEntropy, nil
}

// identifyCommunities identifies which community each DID belongs to
func (g *NetworkGraphAnalyzer) identifyCommunities(ctx context.Context, vouches []*VouchData, context string, epoch int64) (map[string]string, error) {
	// Simplified community detection - in practice would use more sophisticated algorithms
	// like Louvain method, modularity optimization, etc.
	
	communities := make(map[string]string)
	
	// Simple clustering based on mutual vouching patterns
	for _, vouch := range vouches {
		if _, exists := communities[vouch.FromDID]; !exists {
			// Check if this voucher is part of an existing community
			community := g.findBestCommunity(ctx, vouch.FromDID, context, epoch, communities)
			if community == "" {
				// Create new community
				community = fmt.Sprintf("community_%s", vouch.FromDID[:8])
			}
			communities[vouch.FromDID] = community
		}
	}
	
	return communities, nil
}

// findBestCommunity finds the best community for a DID based on connections
func (g *NetworkGraphAnalyzer) findBestCommunity(ctx context.Context, did, context string, epoch int64, communities map[string]string) string {
	// Get vouches from this DID
	outgoingVouches, err := g.getOutgoingVouches(ctx, did, context, epoch)
	if err != nil {
		return ""
	}
	
	// Count connections to existing communities
	communityConnections := make(map[string]int)
	for _, vouch := range outgoingVouches {
		if community, exists := communities[vouch.ToDID]; exists {
			communityConnections[community]++
		}
	}
	
	// Return community with most connections
	bestCommunity := ""
	maxConnections := 0
	for community, connections := range communityConnections {
		if connections > maxConnections {
			maxConnections = connections
			bestCommunity = community
		}
	}
	
	return bestCommunity
}

// getOutgoingVouches gets vouches made by a DID
func (g *NetworkGraphAnalyzer) getOutgoingVouches(ctx context.Context, fromDID, context string, epoch int64) ([]*VouchData, error) {
	// This would typically query the data provider for vouches WHERE from_did = fromDID
	// For now, return empty slice as this requires additional data provider methods
	return []*VouchData{}, nil
}

// GetCommunityOverlap implements GraphAnalyzer.GetCommunityOverlap
func (g *NetworkGraphAnalyzer) GetCommunityOverlap(ctx context.Context, did, context string, epoch int64) (float64, error) {
	vouches, err := g.dataProvider.GetVouches(ctx, did, context, epoch)
	if err != nil {
		return 0.0, err
	}
	
	if len(vouches) <= 1 {
		return 0.0, nil // No overlap possible
	}
	
	communities, err := g.identifyCommunities(ctx, vouches, context, epoch)
	if err != nil {
		return 0.0, err
	}
	
	// Calculate overlap coefficient between communities
	totalOverlap := 0.0
	comparisons := 0
	
	communityMembers := make(map[string][]string)
	for voucher, community := range communities {
		communityMembers[community] = append(communityMembers[community], voucher)
	}
	
	// Compare each pair of communities
	communityList := make([]string, 0, len(communityMembers))
	for community := range communityMembers {
		communityList = append(communityList, community)
	}
	
	for i := 0; i < len(communityList); i++ {
		for j := i + 1; j < len(communityList); j++ {
			overlap := g.calculateCommunityOverlap(communityMembers[communityList[i]], communityMembers[communityList[j]])
			totalOverlap += overlap
			comparisons++
		}
	}
	
	if comparisons == 0 {
		return 0.0, nil
	}
	
	return totalOverlap / float64(comparisons), nil
}

// calculateCommunityOverlap calculates overlap between two communities using Jaccard coefficient
func (g *NetworkGraphAnalyzer) calculateCommunityOverlap(community1, community2 []string) float64 {
	// Convert to sets for easier intersection/union operations
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, member := range community1 {
		set1[member] = true
	}
	for _, member := range community2 {
		set2[member] = true
	}
	
	// Calculate intersection
	intersection := 0
	for member := range set1 {
		if set2[member] {
			intersection++
		}
	}
	
	// Calculate union
	union := len(set1) + len(set2) - intersection
	
	if union == 0 {
		return 0.0
	}
	
	// Jaccard coefficient
	return float64(intersection) / float64(union)
}

// GetDenseSubgraphs implements GraphAnalyzer.GetDenseSubgraphs
func (g *NetworkGraphAnalyzer) GetDenseSubgraphs(ctx context.Context, context string, epoch int64, threshold float64) ([]*DenseSubgraph, error) {
	graph, err := g.buildVouchGraph(ctx, context, epoch)
	if err != nil {
		return nil, fmt.Errorf("failed to build vouch graph: %w", err)
	}
	
	return g.findDenseSubgraphs(graph, threshold)
}

// findDenseSubgraphs finds dense subgraphs using a greedy approach
func (g *NetworkGraphAnalyzer) findDenseSubgraphs(graph *VouchGraph, densityThreshold float64) ([]*DenseSubgraph, error) {
	var denseSubgraphs []*DenseSubgraph
	visited := make(map[string]bool)
	
	// Convert nodes map to slice for iteration
	nodes := make([]string, 0, len(graph.Nodes))
	for node := range graph.Nodes {
		nodes = append(nodes, node)
	}
	
	// Find dense subgraphs starting from each unvisited node
	for _, node := range nodes {
		if visited[node] {
			continue
		}
		
		subgraph := g.growDenseSubgraph(graph, node, densityThreshold, visited)
		if subgraph != nil && len(subgraph.Nodes) >= 3 {
			denseSubgraphs = append(denseSubgraphs, subgraph)
		}
	}
	
	return denseSubgraphs, nil
}

// growDenseSubgraph grows a dense subgraph starting from a seed node
func (g *NetworkGraphAnalyzer) growDenseSubgraph(graph *VouchGraph, seedNode string, threshold float64, visited map[string]bool) *DenseSubgraph {
	subgraphNodes := []string{seedNode}
	subgraphEdges := 0
	visited[seedNode] = true
	
	// Greedy growth: add nodes that increase density
	improved := true
	for improved {
		improved = false
		bestCandidate := ""
		bestDensity := 0.0
		
		// Consider all neighbors of current subgraph
		candidates := g.getNeighbors(graph, subgraphNodes, visited)
		
		for _, candidate := range candidates {
			// Calculate density if we add this candidate
			newNodes := append(subgraphNodes, candidate)
			newEdges := g.countEdgesInSubgraph(graph, newNodes)
			newDensity := g.calculateDensity(len(newNodes), newEdges)
			
			if newDensity > bestDensity && newDensity >= threshold {
				bestDensity = newDensity
				bestCandidate = candidate
				improved = true
			}
		}
		
		if improved {
			subgraphNodes = append(subgraphNodes, bestCandidate)
			subgraphEdges = g.countEdgesInSubgraph(graph, subgraphNodes)
			visited[bestCandidate] = true
		}
	}
	
	density := g.calculateDensity(len(subgraphNodes), subgraphEdges)
	if density < threshold {
		return nil
	}
	
	return &DenseSubgraph{
		Context:   "", // Set by caller
		Epoch:     0,  // Set by caller
		Nodes:     subgraphNodes,
		Edges:     subgraphEdges,
		Density:   density,
		Suspicion: g.calculateSuspicionScore(subgraphNodes, subgraphEdges, density),
	}
}

// getNeighbors returns unvisited neighbors of the given nodes
func (g *NetworkGraphAnalyzer) getNeighbors(graph *VouchGraph, nodes []string, visited map[string]bool) []string {
	neighborSet := make(map[string]bool)
	
	for _, node := range nodes {
		if edges, exists := graph.AdjList[node]; exists {
			for _, edge := range edges {
				if !visited[edge.To] && !neighborSet[edge.To] {
					neighborSet[edge.To] = true
				}
				if !visited[edge.From] && !neighborSet[edge.From] {
					neighborSet[edge.From] = true
				}
			}
		}
	}
	
	neighbors := make([]string, 0, len(neighborSet))
	for neighbor := range neighborSet {
		neighbors = append(neighbors, neighbor)
	}
	
	return neighbors
}

// countEdgesInSubgraph counts edges within a subgraph
func (g *NetworkGraphAnalyzer) countEdgesInSubgraph(graph *VouchGraph, nodes []string) int {
	nodeSet := make(map[string]bool)
	for _, node := range nodes {
		nodeSet[node] = true
	}
	
	edgeCount := 0
	for _, edge := range graph.Edges {
		if nodeSet[edge.From] && nodeSet[edge.To] {
			edgeCount++
		}
	}
	
	return edgeCount
}

// calculateDensity calculates graph density: 2*edges / (nodes*(nodes-1))
func (g *NetworkGraphAnalyzer) calculateDensity(nodeCount, edgeCount int) float64 {
	if nodeCount <= 1 {
		return 0.0
	}
	
	maxPossibleEdges := nodeCount * (nodeCount - 1) / 2
	if maxPossibleEdges == 0 {
		return 0.0
	}
	
	return float64(edgeCount) / float64(maxPossibleEdges)
}

// calculateSuspicionScore calculates how suspicious a dense subgraph is
func (g *NetworkGraphAnalyzer) calculateSuspicionScore(nodes []string, edges int, density float64) float64 {
	// Base suspicion on density and size
	sizeFactor := math.Log(float64(len(nodes))) / 10.0  // Larger groups more suspicious
	densityFactor := density                            // Higher density more suspicious
	
	suspicion := (sizeFactor + densityFactor) / 2.0
	
	// Bound to [0, 1]
	if suspicion > 1.0 {
		suspicion = 1.0
	}
	if suspicion < 0.0 {
		suspicion = 0.0
	}
	
	return suspicion
}

// calculateVouchVolume calculates total vouch volume within a group
func (g *NetworkGraphAnalyzer) calculateVouchVolume(nodes []string, graph *VouchGraph) float64 {
	nodeSet := make(map[string]bool)
	for _, node := range nodes {
		nodeSet[node] = true
	}
	
	totalVolume := 0.0
	for _, edge := range graph.Edges {
		if nodeSet[edge.From] && nodeSet[edge.To] {
			totalVolume += edge.Weight
		}
	}
	
	return totalVolume
}

// calculateCollusionConfidence calculates confidence in collusion detection
func (g *NetworkGraphAnalyzer) calculateCollusionConfidence(subgraph *DenseSubgraph) float64 {
	// Base confidence on multiple factors
	densityConfidence := subgraph.Density           // Higher density = higher confidence
	sizeConfidence := math.Min(1.0, float64(len(subgraph.Nodes))/10.0) // Larger groups up to 10
	
	// Average the confidence factors
	confidence := (densityConfidence + sizeConfidence) / 2.0
	
	// Apply suspicion factor
	confidence = (confidence + subgraph.Suspicion) / 2.0
	
	return confidence
}