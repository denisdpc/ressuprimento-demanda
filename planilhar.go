package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/dustin/go-humanize"
)

type Item struct {
	qtd          int64
	nomenclatura string
	referencia   map[string]Referencia // Nº requisição --> Referencia
}

type Referencia struct {
	requisicao string
	status     string
	data       string
	qtd        int64
	unidade    string
	valor      float64
	correcao   float64
}

var items map[string]*Item

func identificarArquivoReduzido() string { // nome do arquivo reduzido mais recente do diretório planilhas
	files, err := ioutil.ReadDir("./planilhas")
	if err != nil {
		fmt.Println("não foi localizada a planilha reduzida no diretório \"planilhas\"")
	}

	dtMax := ""
	for _, f := range files {
		dt := ""
		arqNome := f.Name()
		if strings.HasSuffix(arqNome, "_reduzido.csv") {
			dt = arqNome[9:21]
			if dt > dtMax {
				dtMax = dt
			}
		}
	}

	return "./planilhas/" + "PLJ0461P_" + dtMax + "_reduzido.csv"
}

func identificarArquivoReferencia() string { // nome do arquivo de referência mais recente do diretório histórico
	files, err := ioutil.ReadDir("./historico")
	if err != nil {
		fmt.Println("não foi localizada a planilha referencia no diretório \"historico\"")
	}

	dtMax := ""
	for _, f := range files {
		dt := ""
		arqNome := f.Name()
		if strings.HasPrefix(arqNome, "referencia") {
			dt = arqNome[10:18]
			if dt > dtMax {
				dtMax = dt
			}
		}
	}

	return "./historico/" + "referencia" + dtMax + ".csv"
}

// carrega o mapa items com os dados do arquivo reduzido
func carregarItems(arqReduzido string) {
	csvArq, err := os.Open(arqReduzido)
	if err != nil {
		fmt.Println("não é possível abrir o arquivo", err)
	}
	defer csvArq.Close()

	rArq := csv.NewReader(csvArq)
	rArq.Comma = ';'

	items = make(map[string]*Item)

	for {
		r, err := rArq.Read()
		if err == io.EOF {
			break
		}
		aux, _ := strconv.ParseInt(r[2], 10, 64)
		items[r[0]] = &Item{
			nomenclatura: r[1],
			qtd:          aux,
		}
	}
}

// carrega o mapa item/RequiçõesRef com os dados do arquivo referencia
func carregarReferencia(arqReferencia string) {
	csvArq, err := os.Open(arqReferencia)
	if err != nil {
		fmt.Println("não é possível abrir o arquivo", err)
	}
	defer csvArq.Close()

	rArq := csv.NewReader(csvArq)
	rArq.Comma = ';'

	for {
		r, err := rArq.Read()
		if err == io.EOF {
			break
		}
		partNumber := r[0]

		if item, temItem := items[partNumber]; temItem {
			auxQtd, _ := strconv.ParseInt(r[4], 10, 64)
			auxValor, _ := strconv.ParseFloat(strings.Replace(r[6], ",", ".", 1), 64)
			auxCorrecao, _ := strconv.ParseFloat(strings.Replace(r[7], ",", ".", 1), 64)
			req := r[1]
			ref := Referencia{
				requisicao: r[1],
				status:     r[2],
				data:       r[3],
				qtd:        auxQtd,
				unidade:    r[5],
				valor:      auxValor,
				correcao:   auxCorrecao,
			}

			if item.referencia == nil {
				item.referencia = make(map[string]Referencia)
			}
			item.referencia[req] = ref
		}
	}
}

/*
disponibiliza uma lista em ordem alfabética crescente de DCN
e permite a escolha daqueles de interesse através de um fator
indexador para cada arquivo.
O retorno corresponde aos partNumbers dos DCN escolhidos
*/
func escolherItems() []string {
	chaves := make([]string, 0, len(items))
	for k := range items {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves)

	indexPN := make(map[int64]string)
	var cont int64 = 0
	for _, partNumber := range chaves {
		cont++
		indexPN[cont] = partNumber
		fmt.Println(cont, partNumber)
	}
	reader := bufio.NewReader(os.Stdin)
	escolhidos, _ := reader.ReadString('\n')

	aux := strings.Split(escolhidos[0:len(escolhidos)-2], ",")
	pnEscolhidos := make([]string, 0, len(aux))

	for _, i := range aux {
		index, _ := strconv.ParseInt(i, 10, 64)
		pnEscolhidos = append(pnEscolhidos, indexPN[index])
	}

	return pnEscolhidos
}

type ByData []Referencia

func (a ByData) Len() int           { return len(a) }
func (a ByData) Less(i, j int) bool { return a[i].data > a[j].data }
func (a ByData) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

/*
disponibiliza uma lista das requisições de referencia
para escolha através de um fator indexador.
O retorno corresponde ao número da requisição escolhida
*/
func escolherReferencia(partNumber string) string {
	item := items[partNumber]

	if item == nil {
		return ""
	}

	requisicoes := make([]Referencia, 0, len(item.referencia))
	for _, ref := range item.referencia {
		requisicoes = append(requisicoes, ref)
	}
	sort.Sort(ByData(requisicoes))

	fmt.Println("------------------------")
	fmt.Println(partNumber, "(", item.qtd, ") : "+item.nomenclatura)

	reqIndex := make(map[int64]string) // index --> nº requisição
	for i, ref := range requisicoes {
		fmt.Println(i, ref.data, ref.requisicao,
			lpad(strconv.FormatInt(ref.qtd, 10), " ", 6), ref.unidade,
			lpad(humanize.CommafWithDigits(ref.valor*ref.correcao, 2), " ", 12))
		reqIndex[int64(i)] = ref.requisicao
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Requisição: ")
	reqEscolhida, _ := reader.ReadString('\n')

	aux, _ := strconv.ParseInt(reqEscolhida[0:len(reqEscolhida)-2], 10, 64)

	return reqIndex[aux]
}

func lpad(s string, pad string, plength int) string {
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}

/*
gera planilha do partNumber considerando
as requisições de referencia escolhidas
*/
func gerarPlanilha(partNumber, requisicaoRef string) {
	fmt.Println("PLANILHA:", partNumber, requisicaoRef)
	path := "./estimativa/"

	f, err := excelize.OpenFile(path + "modelo.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	cell, err := f.GetCellValue("planilha", "A1")
	if err != nil {
		fmt.Println(err)
		return
	}
	f.SetCellValue("planilha", "A3", 10)

	if err := f.SaveAs(path + partNumber + ".xlsx"); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cell)

}

func main() {

	arqReduzido := identificarArquivoReduzido()
	fmt.Println(arqReduzido)

	arqReferencia := identificarArquivoReferencia()
	fmt.Println(arqReferencia)

	carregarItems(arqReduzido)
	carregarReferencia(arqReferencia)

	fmt.Println("LISTAGEM DE ITENS:")
	itensEscolhidos := escolherItems()
	for _, partNumber := range itensEscolhidos {
		requisicaoRef := escolherReferencia(partNumber)
		gerarPlanilha(partNumber, requisicaoRef)
	}

}
