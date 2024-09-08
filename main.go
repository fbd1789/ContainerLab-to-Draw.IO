package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"flag"

	"gopkg.in/yaml.v2"
)

type LabConfig struct {
	Name    string `yaml:"name"`
	Topology struct {
		Nodes map[string]Node `yaml:"nodes"`
		Links []Link          `yaml:"links"`
	} `yaml:"topology"`
}

type Node struct {
	Kind     string `yaml:"kind"`
	MgmtIPv4 string `yaml:"mgmt-ipv4"`
	Env      map[string]string
	Binds    []string
}

type Link struct {
	Endpoints []string `yaml:"endpoints"`
}

// Parse YAML file and map it to Go structs
func parseYAML(filePath string) (*LabConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config LabConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Count connections for each node based on the links
func countConnections(links []Link) map[string]int {
	connectionCount := make(map[string]int)
	for _, link := range links {
		for _, endpoint := range link.Endpoints {
			node := strings.Split(endpoint, ":")[0]
			connectionCount[node]++
		}
	}
	return connectionCount
}

// Identify spines based on the number of connections
func identifySpines(connectionCount map[string]int) []string {
	var spines []string
	for node, count := range connectionCount {
		if count > 4 { // Arbitrary threshold for determining spines
			spines = append(spines, node)
		}
	}
	return spines
}

// Generate Draw.io XML (corrected format for diagram generation)
func generateDrawIO(nodes map[string]Node, links []Link, spines []string) string {
    var sb strings.Builder

    // Start Draw.io XML format
    sb.WriteString("<mxGraphModel><root>\n")
    sb.WriteString("<mxCell id=\"0\"/>\n<mxCell id=\"1\" parent=\"0\"/>\n")

    // Positions for spines and leafs
    spinePosX := 100
    spinePosY := 100
    leafPosX := 100
    leafPosY := 300

    // Add spines to the diagram
    for nodeName := range nodes {
        if contains(spines, nodeName) {
			// find all the connection for the node
			connnectionPoints := extractLinkPerNode(links,nodeName)
            // sb.WriteString(fmt.Sprintf("<mxCell id=\"%s\" value=\"%s (Spine);\" style=\"rounded=1;whiteSpace=wrap;html=1;strokeColor=#000000;fillColor=#FFFFFF;shape=Rectangle;autosize=1;\" vertex=\"1;\" parent=\"1\">\n", nodeName, nodeName))
            // sb.WriteString(fmt.Sprintf("<mxGeometry x=\"%d\" y=\"%d\" width=\"80\" height=\"30\" as=\"geometry\"/>\n", spinePosX, spinePosY))
            // sb.WriteString("</mxCell>\n")
            // spinePosX += 150 // Space out the spines horizontally

			sb.WriteString(fmt.Sprintf("<UserObject label=\"%s\" tooltip=\"%s\" id=\"%s\">\n", nodeName, connnectionPoints,nodeName))
			sb.WriteString("<mxCell style=\"rounded=0;whiteSpace=wrap;html=1;\" vertex=\"1\" parent=\"1\">\n")
			sb.WriteString(fmt.Sprintf("<mxGeometry x=\"%d\" y=\"%d\" width=\"80\" height=\"30\" as=\"geometry\" />\n", spinePosX, spinePosY))
			sb.WriteString("</mxCell>\n")
			sb.WriteString("</UserObject>\n")
			spinePosX += 150 // Space out the spines horizontally

        }
    }

	// Find mlag node
	var mlag []string
	for _, link := range links {
		// Extract the node names from both ends of the link
		endpoints := link.Endpoints
		nodeName0 := strings.Split(endpoints[0], ":")[0]
		nodeName1 := strings.Split(endpoints[1], ":")[0]
	
		// Check that neither nodeName0 nor nodeName1 are in spines
		if !contains(spines, nodeName0) && !contains(spines, nodeName1) {
			mlag = append(mlag, nodeName0, nodeName1)
		}
	}

	// Add leafs in mlag to the diagram
	for _,nodeName := range mlag {
        if !contains(spines, nodeName) {
            sb.WriteString(fmt.Sprintf("<mxCell id=\"%s\" value=\"%s (Leaf)\" style=\"rounded=1;whiteSpace=wrap;html=1;\" vertex=\"1\" parent=\"1\">\n", nodeName, nodeName))
            sb.WriteString(fmt.Sprintf("<mxGeometry x=\"%d\" y=\"%d\" width=\"80\" height=\"30\" as=\"geometry\"/>\n", leafPosX, leafPosY))
            sb.WriteString("</mxCell>\n")
            leafPosX += 100 // Space out the leafs horizontally
        }
    }

    // Add leafs to the diagram
    for nodeName := range nodes {
        if !contains(spines, nodeName) && !contains(mlag, nodeName) {
            sb.WriteString(fmt.Sprintf("<mxCell id=\"%s\" value=\"%s (Leaf)\" style=\"rounded=1;whiteSpace=wrap;html=1;\" vertex=\"1\" parent=\"1\">\n", nodeName, nodeName))
            sb.WriteString(fmt.Sprintf("<mxGeometry x=\"%d\" y=\"%d\" width=\"80\" height=\"30\" as=\"geometry\"/>\n", leafPosX, leafPosY))
            sb.WriteString("</mxCell>\n")
            leafPosX += 100 // Space out the leafs horizontally
        }
    }

    // Add links to the diagram
    for _, link := range links {
        endpoint1 := strings.Split(link.Endpoints[0], ":")[0]
        endpoint2 := strings.Split(link.Endpoints[1], ":")[0]
        sb.WriteString(fmt.Sprintf("<mxCell edge=\"1\" parent=\"1\" source=\"%s\" target=\"%s\" style=\"endArrow=none\">\n", endpoint1, endpoint2))
        sb.WriteString("<mxGeometry relative=\"1\" as=\"geometry\"/>\n")
        sb.WriteString("</mxCell>\n")
    }

    // End Draw.io XML format
    sb.WriteString("</root></mxGraphModel>\n")

    return sb.String()
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to find the link information for each node
func extractLinkPerNode (slice []Link, item string) string{
	var sb strings.Builder

    for _, link := range slice {
        endpoints := link.Endpoints
        
        // Vérifier si l'un des endpoints contient le noeud spécifié (item)
        if strings.Contains(endpoints[0], item) {
            nodeInterface := strings.Split(endpoints[0], ":")[1]
            sb.WriteString(fmt.Sprintf("%s to %s&#xa;", nodeInterface, endpoints[1]))
        } else if strings.Contains(endpoints[1], item) {
            nodeInterface := strings.Split(endpoints[1], ":")[1]
            sb.WriteString(fmt.Sprintf("%s to %s&#xa;", nodeInterface, endpoints[0]))
        }
    }

    // Return the result
    return sb.String()
}

func main() {
	// Enter the fileName as a parameter
	var fileName string
	flag.StringVar(&fileName,"f","exampleLab.yml,","The ContainerLab.yaml file")
	flag.Parse()

	// Load and parse the YAML file
	labConfig, err := parseYAML(fileName)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}

	// Count connections for each node
	connectionCount := countConnections(labConfig.Topology.Links)

	// Display the number of connections for each node
	fmt.Println("Connection Count:")
	for node, count := range connectionCount {
		fmt.Printf("  %s: %d\n", node, count)
	}

	// Identify spines based on connections
	spines := identifySpines(connectionCount)

	// Display the identified spines
	fmt.Println("Spines Identified:")
	for _, spine := range spines {
		fmt.Printf("  %s\n", spine)
	}

	// Generate Draw.io XML for the diagram
	drawioXML := generateDrawIO(labConfig.Topology.Nodes, labConfig.Topology.Links, spines)

	// Write the Draw.io XML to a file
	err = os.WriteFile("diagram.drawio", []byte(drawioXML), 0644)
	if err != nil {
		log.Fatalf("Error writing DrawIO file: %v", err)
	}

	fmt.Println("DrawIO diagram generated as 'diagram.drawio'")	
}
