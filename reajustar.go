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

var statusValido = map[string]bool{
	"Mapa Aprovado":                       true,
	"Empenho Aprovado":                    true,
	"Expedida":                            true,
	"Recebida Parcialmente no Solicitant": true,
	"Recebida no Solicitante":             true,
	"Recebida na Comissão":                true,
	"Controle de Qualidade":               true,
	"Volume no Solicitante":               true,
}

func getPlanilhaNome() string {
	files, err := ioutil.ReadDir("./historico")
	if err != nil {
		fmt.Println("a planilha de histórico de requisições deve ser inserida no subdiretório \"historico\"")
	}

	temPlanilhaHistorico := false
	var arqNome string
	for _, f := range files {
		arqNome = f.Name()
		if strings.HasPrefix(arqNome, "PLJ0461P_") {
			temPlanilhaHistorico = true
			break
		}
	}
	if !temPlanilhaHistorico {
		fmt.Println("a planilha de histórico de requisições deve ser inserida no subdiretório \"historico\"")
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
		fmt.Println("desconsiderado:", req.numero, req.status)
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

func gravarPlanilha(tabelaCorrecao map[string]float64) {
	mesPassado := time.Now().AddDate(0, -1, 0).Format("2006-01")
	csvfile, err := os.Create("historico/referencia " + mesPassado + ".csv")

	if err != nil {
		fmt.Println("ERRO:", err)
		return
	}
	defer csvfile.Close()

	w := csv.NewWriter(csvfile)
	w.Comma = ';'

	w.Write([]string{"Part Number", "Req numero", "Req status", "Req data", "qtd", "Unidade", "Valor unitario", "IGPM acumulado"})

	for pn, requisicoes := range reqs {
		for _, req := range requisicoes {
			dt := req.dataInclusao.Format("2006-01")
			correcao := tabelaCorrecao[dt]

			w.Write([]string{
				pn,
				req.numero,
				req.status,
				dt,
				strconv.FormatFloat(req.qtd, 'f', 0, 64),
				req.unidade,
				strings.Replace(strconv.FormatFloat(req.valor, 'f', 2, 64), ".", ",", 1),
				strings.Replace(strconv.FormatFloat(correcao, 'f', 8, 64), ".", ",", 1)})
		}
	}
	w.Flush()
}

func calcularTabelaCorrecao(igpm map[string]float64) map[string]float64 { // DATA --> ACUMULADO
	tabelaCorrecao := make(map[string]float64)
	dataInicial := time.Date(1990, 01, 01, 00, 00, 00, 00, time.UTC) // 1º data de IGPM
	dataRef := time.Now().AddDate(0, -1, 0)                          // mês passado

	dt := dataRef.Format("2006-01")
	tabelaCorrecao[dt] = igpm[dt]
	valorAnterior := tabelaCorrecao[dt]

	for {
		dataRef = dataRef.AddDate(0, -1, 0)
		if dataRef.Before(dataInicial) {
			break
		}
		dt = dataRef.Format("2006-01")

		tabelaCorrecao[dt] = valorAnterior * igpm[dt]
		valorAnterior = tabelaCorrecao[dt]
	}
	return tabelaCorrecao
}

func lerIGPM() map[string]float64 {
	csvfile, err := os.Open("historico/IGPM.csv")
	if err != nil {
		log.Fatalln("não é possível abrir o arquivo IGMP.csv", err)
	}

	r := csv.NewReader(csvfile)
	r.Comma = ';'

	igpm := make(map[string]float64)

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
	}
	mesPassado := time.Now().AddDate(0, -1, 0).Format("2006-01")

	_, igpmRegistrado := igpm[mesPassado]
	if !igpmRegistrado {
		fmt.Println("incluir IGPM de", mesPassado)
		return nil
	}
	return igpm
}

func main() {
	igpm := lerIGPM()
	if igpm != nil {
		tabelaCorrecao := calcularTabelaCorrecao(igpm)
		planilha := getPlanilhaNome()
		lerPlanilha(planilha)
		gravarPlanilha(tabelaCorrecao)
	}
}
