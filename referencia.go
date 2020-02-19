package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	NUMERO = 1
)

type Requisicao struct {
	numero, status, partNumber, unidade string
	qtd                                 float64
	dataInclusao                        time.Time
}

var reqs map[string][]Requisicao // DCN --> Requisicao

var statusConsiderar = map[string]bool{
	"Empenho Aprovado":                    true,
	"Expedida":                            true,
	"Recebida Parcialmente no Solicitant": true,
	"Recebida no Solicitante":             true,
}

func getPlanilhaNome() string {
	files, err := ioutil.ReadDir("./historico")
	if err != nil {
		fmt.Println("as planilhas de requisição devem ser inseridas no subdiretório \"historico\"")
	}

	var arqNome string
	for _, f := range files {
		arqNome = f.Name()
	}

	return arqNome
}

func extrairDadosLinha(linha string) Requisicao {
	col := strings.Split(linha, ";")
	req := Requisicao{}
	req.numero = strings.TrimSpace(col[NUMERO])
	if len(req.numero) == 0 ||
		req.numero == "--------------" ||
		req.numero == "Nº Requisição" {
		req.numero = ""
		return req
	}
	if strings.TrimSpace(col[14]) == "Material Extra-Sistema" {
		req.numero = ""
		return req
	}

	return req
}

func lerPlanilha(arq string) {
	f, err := os.Open("atual/" + arq + ".CSV")
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

func main() {
	planilha := getPlanilhaNome()
	fmt.Println(planilha)
}
