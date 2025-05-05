package main

import (
	"log"
	"strings"
	"unicode"

	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

// vocab é Vocabulario. Define que tokens considerar e ignorar.
var vocab = []string{
	"ligar", "desligar", "acender", "apagar",
	"luz", "ventilador", "ar", "condicionado",
	"como", "está", "temperatura", "hoje", "agora",
}

// classNames são as classes que desejamos utilizar para classificar sentenças.
var classNames = []string{
	"comando_ligar",
	"comando_desligar",
	"comando_consulta",
}

// trainSentences são sentenças que utilizaremos para treinar a rede.
var trainSentences = []string{
	"ligar a luz",
	"acender o ventilador",
	"ligar o ar condicionado",
	"desligar a luz",
	"apagar o ventilador",
	"desligar o ar",
	"como está a temperatura",
	"qual é a temperatura agora",
	"me diga a temperatura de hoje",
}

// trainLabels - etiquetas para classificarmos nossas sentenças de treinamento.
var trainLabels = []int{
	0, 0, 0,
	1, 1, 1,
	2, 2, 2,
}

// wordToIndex é o indexador dos tokens.
var wordToIndex = func() map[string]int {
	m := make(map[string]int)
	for i, word := range vocab {
		m[word] = i
	}

	return m
}()

// vectorize: função simples para demonstrar a vetorização.
func vectorize(text string) []float32 {
	vec := make([]float32, len(vocab))
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	for _, word := range words {
		if idx, ok := wordToIndex[word]; ok {
			vec[idx]++
		}
	}

	return vec
}

// argmax busca o id do maior valor na slice fornecida.
func argmax(vals []float32) int {
	maxIdx := 0
	maxVal := vals[0]

	for i, v := range vals {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	return maxIdx
}

// train: aqui treinamos nossa rede, e posteriormente retornamos
//
//	B: nó com os viezes (biases)
//	W: nó com os pesos para análise.
func train() (biases *gorgonia.Node, weights *gorgonia.Node, err error) {
	numClasses := len(classNames)
	numSamples := len(trainSentences)
	inputSize := len(vocab)

	// Build input and target tensors
	Xdata := make([]float32, 0, numSamples*inputSize)
	Ydata := make([]float32, 0, numSamples*numClasses)

	for i, sentence := range trainSentences {
		Xdata = append(Xdata, vectorize(sentence)...)

		for j := range numClasses {
			switch {
			case j == trainLabels[i]:
				Ydata = append(Ydata, 1)
			default:
				Ydata = append(Ydata, 0)
			}
		}
	}

	// tensores são estruturas de dados multidimensionais que carregam inputs e outputs em uma rede neural.
	Xtensor := tensor.New(tensor.WithShape(numSamples, inputSize), tensor.WithBacking(Xdata))
	Ytensor := tensor.New(tensor.WithShape(numSamples, numClasses), tensor.WithBacking(Ydata))

	// o grafo é a forma de modelar o conjunto de instruções, é o "programa" da rede neural.
	g := gorgonia.NewGraph()

	// inputs
	X := gorgonia.NewMatrix(g, gorgonia.Float32, gorgonia.WithShape(numSamples, inputSize), gorgonia.WithName("X"))

	// outputs
	Y := gorgonia.NewMatrix(g, gorgonia.Float32, gorgonia.WithShape(numSamples, numClasses), gorgonia.WithName("Y"))

	// weights é onde armazenaremos os pesos uma vez que a rede esteja treinada - compartilharemos essa matriz
	// com a etapa de predição.
	weights = gorgonia.NewMatrix(g, gorgonia.Float32, gorgonia.WithShape(inputSize, numClasses), gorgonia.WithInit(gorgonia.GlorotN(1)))

	// de forma análoga, armazenaremos os viezes na matriz abaixo.
	biases = gorgonia.NewVector(g, gorgonia.Float32, gorgonia.WithShape(numClasses), gorgonia.WithInit(gorgonia.Zeroes()))

	// valores brutos, não normalizados.
	logits := gorgonia.Must(gorgonia.BroadcastAdd(gorgonia.Must(gorgonia.Mul(X, weights)), biases, nil, []byte{0}))

	// valores normalizados.
	pred := gorgonia.Must(gorgonia.SoftMax(logits))

	// Calculo da perda.
	logPred := gorgonia.Must(gorgonia.Log(pred))
	mul := gorgonia.Must(gorgonia.HadamardProd(Y, logPred))
	sum := gorgonia.Must(gorgonia.Sum(mul))
	neg := gorgonia.Must(gorgonia.Neg(sum))

	// loss = -mean(y * log(pred))
	// loss é a perda, ou a distância da resposta dada em relação ao resultado esperado
	// menor perda => melhor resultado.
	loss := gorgonia.Must(gorgonia.Div(neg, gorgonia.NewConstant(float32(numSamples))))

	// Gradientes: indicam ao otimizador como ajustar os pesos para reduzir a perda.
	if _, err = gorgonia.Grad(loss, weights, biases); err != nil {
		return nil, nil, err
	}

	vm := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(weights, biases))

	// vamos usar o otimizador ADAM - ele é simples e já vem implementado.
	solver := gorgonia.NewAdamSolver()

	// Treinamento
	// Como ja deixamos tudo preparado, a cada passo, a perda é avaliada e
	// o otimizador recalibra os pesos.
	for range 3000 {
		if err = gorgonia.Let(X, Xtensor); err != nil {
			return nil, nil, err
		}

		if err = gorgonia.Let(Y, Ytensor); err != nil {
			return nil, nil, err
		}

		if err = vm.RunAll(); err != nil {
			return nil, nil, err
		}

		if err = solver.Step([]gorgonia.ValueGrad{weights, biases}); err != nil {
			return nil, nil, err
		}

		vm.Reset()
	}

	return biases, weights, nil
}

func predict(sentences []string, biases *gorgonia.Node, weights *gorgonia.Node) error {
	inputSize := len(vocab)
	numClasses := len(classNames)

	g2 := gorgonia.NewGraph()
	X2 := gorgonia.NewMatrix(g2, gorgonia.Float32, gorgonia.WithShape(1, inputSize), gorgonia.WithName("X2"))
	W2 := gorgonia.NewMatrix(g2, gorgonia.Float32, gorgonia.WithShape(inputSize, numClasses), gorgonia.WithValue(weights.Value()))
	B2 := gorgonia.NewVector(g2, gorgonia.Float32, gorgonia.WithShape(numClasses), gorgonia.WithValue(biases.Value()))

	out := gorgonia.Must(gorgonia.SoftMax(gorgonia.Must(gorgonia.Add(gorgonia.Must(gorgonia.Mul(X2, W2)), B2))))

	for _, sentence := range sentences {
		vec := vectorize(sentence)
		input := tensor.New(tensor.WithShape(1, inputSize), tensor.WithBacking(vec))

		if err := gorgonia.Let(X2, input); err != nil {
			return err
		}

		machine := gorgonia.NewTapeMachine(g2)
		if err := machine.RunAll(); err != nil {
			return err
		}

		probs, ok := out.Value().Data().([]float32)
		if !ok {
			panic("failed to compute probabilities")
		}

		predicted := argmax(probs)

		log.Printf("Input: %-30s  Predicted: %s (%v)\n", sentence, classNames[predicted], probs)
	}

	return nil
}

func main() {
	b, w, err := train()
	if err != nil {
		log.Fatal(err)
	}

	if err = predict(trainSentences, b, w); err != nil {
		log.Fatal(err)
	}

	log.Println("====")

	if err = predict([]string{
		"pode ligar a luz?",          // comando_ligar
		"acenda o ar condicionado",   // comando_ligar
		"desligue o ventilador",      // comando_desligar
		"apague o ar",                // comando_desligar
		"qual a temperatura agora?",  // comando_consulta
		"está quente hoje?",          // comando_consulta
		"ligue o ventilador da sala", // comando_ligar
		"desligar luz do quarto",     // comando_desligar
		"me diga como está o clima",  // comando_consulta
	}, b, w); err != nil {
		log.Fatal(err)
	}
}
