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
)

func comparar(arqAntigo, arqRecente string) {
	if strings.Index(arqAntigo, "reduzido") == -1 ||
		strings.Index(arqRecente, "reduzido") == -1 {
		fmt.Println("arquivo inexistente !")
		return
	}

	csvRecente, err := os.Open("planilhas/" + arqRecente)
	if err != nil {
		log.Fatal("não é possível abrir o arquivo", err)
	}
	defer csvRecente.Close()

	csvAntigo, err := os.Open("planilhas/" + arqAntigo)
	if err != nil {
		log.Fatal("não é possível abrir o arquivo", err)
	}
	defer csvAntigo.Close()

	rRecente := csv.NewReader(csvRecente)
	rRecente.Comma = ';'

	var contRecente int
	qtdRecente := make(map[string]int64)
	for {
		record, err := rRecente.Read()
		if err == io.EOF {
			break
		}
		qtd, _ := strconv.ParseInt(strings.TrimSpace(record[2]), 10, 64)
		qtdRecente[record[0]] = qtd
		contRecente++
	}

	rAntigo := csv.NewReader(csvAntigo)
	rAntigo.Comma = ';'

	var contAntigo int
	qtdAntigo := make(map[string]int64)
	for {
		record, err := rAntigo.Read()
		if err == io.EOF {
			break
		}
		qtd, _ := strconv.ParseInt(strings.TrimSpace(record[2]), 10, 64)
		qtdAntigo[record[0]] = qtd
		contAntigo++
	}

	fmt.Println("(antigo: ", contAntigo, ")  (recente: ", contRecente, ")")
	fmt.Println()

	var chavesDif []string
	for k, vRecente := range qtdRecente {
		if qtdAntigo[k] != vRecente {
			fmt.Println(k, qtdAntigo[k], "==>", vRecente)
			chavesDif = append(chavesDif, k)
		}
	}

	fmt.Println()
	for k, vAntigo := range qtdAntigo {
		if qtdRecente[k] != vAntigo {
			exibirDif := true // se k não está em chavesDif, imprime
			for _, chave := range chavesDif {
				if k == chave {
					exibirDif = false
					break
				}
			}
			if exibirDif {
				fmt.Println(k, vAntigo, "==>", qtdRecente[k])
			}
		}
	}

}

func printDataArquivos(arqAntigoNome, arqRecenteNome string) {
	rAnt := []rune(arqAntigoNome)[9:17]
	rRec := []rune(arqRecenteNome)[9:17]

	fmt.Println(string(rAnt[0:4]) + "-" + string(rAnt[4:6]) + "-" + string(rAnt[6:8]) + " ==> " +
		string(rRec[0:4]) + "-" + string(rRec[4:6]) + "-" + string(rRec[6:8]))
}

func main() {

	files, err := ioutil.ReadDir("./planilhas")
	if err != nil {
		fmt.Println("as planilhas de requisição devem ser inseridas no subdiretório \"planilhas\"")
		return
	}

	for i, f := range files {
		arqNome := f.Name()
		if strings.Index(arqNome, "reduzido") > -1 {
			fmt.Println(i, ":", arqNome)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Arquivo mais antigo: ")
	arqAntigo, _ := reader.ReadString('\n')

	fmt.Print("Arquivo mais recente: ")
	arqRecente, _ := reader.ReadString('\n')

	fmt.Println()

	numArqAntigo, _ := strconv.ParseInt(strings.TrimSpace(arqAntigo), 10, 64)
	numArqRecente, _ := strconv.ParseInt(strings.TrimSpace(arqRecente), 10, 64)

	arqAntigoNome := files[numArqAntigo].Name()
	arqRecenteNome := files[numArqRecente].Name()

	printDataArquivos(arqAntigoNome, arqRecenteNome)

	comparar(arqAntigoNome, arqRecenteNome)
	fmt.Println()
}
