package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Requisicao struct {
	numero, status, partNumber, nomenclatura, unidade string
	qtd                                               float64
}

var statusConsiderar = map[string]bool{
	"Aguardando validação":     true,
	"Validada":                 true,
	"Análise do Pedido":        true,
	"Selecionada para cotação": true,
	"Recotada":                 true,
	"Em Cotação":               true,
	"Suspensa temporariamente": true,
	"Item Deserto":             true,
}

/*
var statusDesconsiderar = map[string]bool{
	"Anulada":                 true,
	"Cancelada":               true,
	"Recebida no Solicitante": true,
	"Recebida na Comissão":    true,
}
*/

var reqs map[string][]Requisicao // partNumber --> Requisicao

var requisicoesDesconsideradas [][]string

// os planilhas são processadas e armazenadas
// com o mesmo nome acrescido de "_reduzido"
func gravarPlanilha(arq string) {
	csvfile, _ := os.Create("planilhas/" + arq + "_reduzido.csv")
	w := csv.NewWriter(csvfile)
	w.Comma = ';'

	for pn, requisicoes := range reqs {
		var record []string
		//var soma float64
		record = append(record, pn)
		record = append(record, requisicoes[0].nomenclatura)
		//record = append(record, "0") // será atualizado com o valor de soma

		for _, req := range requisicoes {
			record = append(record, req.numero)
			record = append(record, strconv.FormatFloat(req.qtd, 'f', 0, 64))
			record = append(record, req.unidade)
			//soma += req.qtd
		}

		//record[2] = strconv.FormatFloat(soma, 'f', 0, 64)
		w.Write(record)
	}
	w.Flush()
	csvfile.Close()
}

func extrairDadosLinha(linha string) Requisicao {
	col := strings.Split(linha, ";")
	req := Requisicao{}
	req.numero = strings.TrimSpace(col[1])
	if len(req.numero) == 0 ||
		req.numero == "--------------" ||
		req.numero == "Nº Requisição" {
		req.numero = ""
		return req
	}

	req.partNumber = strings.TrimSpace(col[4])
	req.status = strings.TrimSpace(col[17])
	if !statusConsiderar[req.status] {
		requisicoesDesconsideradas = append(requisicoesDesconsideradas, []string{req.status, req.partNumber, req.numero})
		req.numero = ""
		return req
	}

	if strings.TrimSpace(col[14]) == "Material Extra-Sistema" {
		req.numero = ""
		return req
	}

	req.partNumber = strings.TrimSpace(col[4])
	req.nomenclatura = strings.TrimSpace(col[7])
	req.unidade = strings.TrimSpace(col[31])
	req.qtd, _ = strconv.ParseFloat(
		strings.ReplaceAll(strings.TrimSpace(col[30]), "\"", ""), 64)

	cff := strings.TrimSpace(col[6])
	if cff != "002FK" {
		fmt.Println("CFF diferente de 002FK: ", cff, " : ", req.numero, " : ", req.partNumber)
	}

	return req
}

func lerPlanilha(arq string) {
	f, err := os.Open("planilhas/" + arq + ".CSV")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	reqs = make(map[string][]Requisicao)

	dec := transform.NewReader(f, charmap.ISO8859_1.NewDecoder())
	scanner := bufio.NewScanner(dec)
	for scanner.Scan() {
		linha := scanner.Text()
		var req Requisicao = extrairDadosLinha(linha)
		if len(req.numero) != 11 {
			continue
		}
		reqs[req.partNumber] = append(reqs[req.partNumber], req)
	}
}

// listar nome de planilhas não processadas
func getPlanilhasNome() []string {
	files, err := ioutil.ReadDir("./planilhas")
	if err != nil {
		fmt.Println("as planilhas de requisição devem ser inseridas no subdiretório \"planilhas\"")
	}

	var arquivos = make(map[string]string)
	for _, f := range files {
		arqNome := f.Name()
		if strings.Index(arqNome, "reduzido") > -1 {
			delete(arquivos, arqNome[0:21])
		} else if strings.HasPrefix(arqNome, "PLJ0461P") {
			arquivos[arqNome[0:21]] = arqNome
		}
	}

	var aux []string
	for _, v := range arquivos {
		aux = append(aux, v[0:21])
	}

	return aux
}

func gravarRequisicoesDesconsideradas(arq string) {
	fmt.Println("DESCONSIDERADAS:")
	sort.Slice(requisicoesDesconsideradas, func(i, j int) bool {
		return requisicoesDesconsideradas[i][0] < requisicoesDesconsideradas[j][0]
	})

	for _, req := range requisicoesDesconsideradas {
		fmt.Println(req)
	}

	csvfile, _ := os.Create("planilhas/" + arq + "_desconsideradas.csv")
	w := csv.NewWriter(csvfile)
	w.Comma = ';'

	for _, requisicao := range requisicoesDesconsideradas {
		w.Write(requisicao)
	}
	w.Flush()
	csvfile.Close()
}

func main() {
	planilhas := getPlanilhasNome()

	for _, arqNome := range planilhas {
		fmt.Println(arqNome)
		lerPlanilha(arqNome)
		gravarPlanilha(arqNome)
		gravarRequisicoesDesconsideradas(arqNome)
	}

}
