package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// === Estructuras ===

type Record struct {
	Attrs []int
	Class int
}

type TreeNode struct {
	Attribute int
	Value     int
	Left      *TreeNode
	Right     *TreeNode
	IsLeaf    bool
	Class     int
}

// === Generador de datos con 5 atributos binarios ===

func generateData(n int, seed int64) []Record {
	rand.Seed(seed)
	var data []Record
	for i := 0; i < n; i++ {
		attrs := make([]int, 5)
		sum := 0
		for j := 0; j < 5; j++ {
			attrs[j] = rand.Intn(2) // 0 o 1
			sum += attrs[j]
		}
		class := 0
		if sum >= 3 {
			class = 1
		}
		data = append(data, Record{Attrs: attrs, Class: class})
	}
	return data
}

// === Cálculo de entropía y ganancia ===

func entropy(records []Record) float64 {
	total := float64(len(records))
	if total == 0 {
		return 0
	}
	count := make(map[int]int)
	for _, r := range records {
		count[r.Class]++
	}
	result := 0.0
	for _, c := range count {
		p := float64(c) / total
		result -= p * math.Log2(p)
	}
	return result
}

func infoGain(records []Record, attr int) float64 {
	total := float64(len(records))
	left, right := split(records, attr, 0)
	entropyBefore := entropy(records)
	entropyAfter := (float64(len(left))/total)*entropy(left) + (float64(len(right))/total)*entropy(right)
	return entropyBefore - entropyAfter
}

// === División y mayoría ===

func split(records []Record, attr, val int) ([]Record, []Record) {
	var left, right []Record
	for _, r := range records {
		if r.Attrs[attr] == val {
			left = append(left, r)
		} else {
			right = append(right, r)
		}
	}
	return left, right
}

func majority(records []Record) int {
	count := make(map[int]int)
	for _, r := range records {
		count[r.Class]++
	}
	if count[0] > count[1] {
		return 0
	}
	return 1
}

// === Ganancia concurrente ===

type GainResult struct {
	Attr int
	Gain float64
}

func bestSplitConcurrent(records []Record) int {
	numAttrs := len(records[0].Attrs)
	gainCh := make(chan GainResult, numAttrs)
	var wg sync.WaitGroup

	for i := 0; i < numAttrs; i++ {
		wg.Add(1)
		go func(attr int) {
			defer wg.Done()
			gain := infoGain(records, attr)
			gainCh <- GainResult{Attr: attr, Gain: gain}
		}(i)
	}

	wg.Wait()
	close(gainCh)

	bestAttr := -1
	bestGain := -1.0

	for result := range gainCh {
		if result.Gain > bestGain {
			bestGain = result.Gain
			bestAttr = result.Attr
		}
	}
	return bestAttr
}

// === Construcción del árbol ===

func buildTree(records []Record, depth int) *TreeNode {
	if len(records) == 0 {
		return nil
	}
	class := records[0].Class
	allSame := true
	for _, r := range records {
		if r.Class != class {
			allSame = false
			break
		}
	}
	if allSame {
		return &TreeNode{IsLeaf: true, Class: class}
	}
	bestAttr := bestSplitConcurrent(records)
	if bestAttr == -1 {
		return &TreeNode{IsLeaf: true, Class: majority(records)}
	}
	left, right := split(records, bestAttr, 0)
	return &TreeNode{
		Attribute: bestAttr,
		Value:     0,
		Left:      buildTree(left, depth+1),
		Right:     buildTree(right, depth+1),
	}
}

// === Clasificación ===

func classify(tree *TreeNode, r Record) int {
	if tree.IsLeaf {
		return tree.Class
	}
	if r.Attrs[tree.Attribute] == tree.Value {
		return classify(tree.Left, r)
	}
	return classify(tree.Right, r)
}

// === Impresión estructurada del árbol ===

func printTree(node *TreeNode, indent string) {
	if node == nil {
		return
	}
	if node.IsLeaf {
		fmt.Printf("%s🟩 Hoja → Clase = %d\n", indent, node.Class)
	} else {
		fmt.Printf("%s🔀 Si atributo[%d] == %d:\n", indent, node.Attribute, node.Value)
		printTree(node.Left, indent+"  ")
		fmt.Printf("%s🔁 Si atributo[%d] != %d:\n", indent, node.Attribute, node.Value)
		printTree(node.Right, indent+"  ")
	}
}

// === Main ===

func main() {
	n := 1000000
	fmt.Printf("Generando %d registros con 5 atributos binarios...\n", n)
	allData := generateData(n, time.Now().UnixNano())

	// División 80/20 entrenamiento/prueba
	splitIndex := int(0.8 * float64(n))
	trainData := allData[:splitIndex]
	testData := allData[splitIndex:]

	fmt.Println("Entrenando árbol de decisión (concurrente)...")
	start := time.Now()
	tree := buildTree(trainData, 0)
	elapsed := time.Since(start)
	fmt.Printf("Árbol entrenado en: %s\n\n", elapsed)

	// Imprimir estructura del árbol (puedes comentar si es muy grande)
	printTree(tree, "")

	// Evaluar en prueba
	correct := 0
	for _, r := range testData {
		if classify(tree, r) == r.Class {
			correct++
		}
	}
	acc := float64(correct) / float64(len(testData)) * 100
	fmt.Printf("\nPrecisión en conjunto de prueba: %.2f%%\n", acc)
}
