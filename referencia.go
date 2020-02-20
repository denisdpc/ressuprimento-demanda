package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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
	qtd, valor                          float64
	dataInclusao                        time.Time
}

var reqs map[string][]Requisicao // DCN --> Requisicao
var igpm map[string]float64

var statusValido = map[string]bool{
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

	req.status = strings.TrimSpace(col[17])
	if !statusValido[req.status] {
		req.numero = ""
		return req
	}

	req.partNumber = strings.TrimSpace(col[4])
	req.unidade = strings.TrimSpace(col[31])
	req.qtd, _ = strconv.ParseFloat(
		strings.ReplaceAll(strings.TrimSpace(col[30]), "\"", ""), 64)
	req.valor, _ = strconv.ParseFloat(
		strings.ReplaceAll(strings.TrimSpace(col[28]), "\"", ""), 64)

	dataAux := []rune(strings.TrimSpace(col[15]))[0:10]
	dataReq, _ := time.Parse("2006-01-02", string(dataAux[6:10])+"-"+string(dataAux[3:5])+"-"+string(dataAux[0:2]))
	req.dataInclusao = dataReq

	return req
}

func lerPlanilha(arq string) {
	f, err := os.Open("historico/" + arq)
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

func calcularCorrecao(dataReq time.Time) float64 {
	hoje := time.Now()
	acumulado := 1.0

	for {
		if dataReq.Year() >= hoje.Year() {
			if dataReq.Month() >= hoje.Month()-1 {
				break
			}
		}
		dataReq = dataReq.AddDate(0, 1, 0)
		dataReqString := dataReq.Format("2006-01")
		indice := igpm[dataReqString]
		acumulado = acumulado * indice
		//fmt.Println(dataReqString, indice)
	}
	fmt.Println(acumulado)

	return acumulado
}

func gravarPlanilha() {
	csvfile, err := os.Create("historico/historico_referencia.csv")
	if err != nil {
		fmt.Println("ERRO:", err)
		return
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	w.Comma = ';'

	for pn, requisicoes := range reqs {
		w.Write([]string{pn})

		fmt.Println(pn)

		for _, req := range requisicoes {
			correcao := calcularCorrecao(req.dataInclusao)
			// TODO: correcao e req.valor trocar . por ,

			w.Write([]string{
				req.numero,
				req.status,
				req.dataInclusao.Format("2006-01"),
				strconv.FormatFloat(req.qtd, 'f', 0, 64),
				strconv.FormatFloat(req.valor, 'f', 2, 64),
				strconv.FormatFloat(correcao, 'f', 8, 64)})

			fmt.Println(req.dataInclusao, req.numero, req.status, req.qtd, req.valor, correcao)
		}
	}
	w.Flush()
}

func lerIGPM() {
	csvfile, err := os.Open("historico/IGPM.csv")
	if err != nil {
		log.Fatalln("não é possível abrir o arquivo IGMP.csv", err)
	}

	r := csv.NewReader(csvfile)
	r.Comma = ';'

	igpm = make(map[string]float64)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		indice, _ := strconv.ParseFloat(record[1], 64)
		igpm[record[0]] = indice
		//dataReq, _ := time.Parse("2006-01-02", string(dataAux[6:10])+"-"+string(dataAux[3:5])+"-"+string(dataAux[0:2]))
		//fmt.Println(record)
	}
	//fmt.Println(igpm)
}

func main() {
	planilha := getPlanilhaNome()
	lerPlanilha(planilha)
	lerIGPM()
	//fmt.Println(planilha)

	gravarPlanilha()
	//lerIGPM()

}
