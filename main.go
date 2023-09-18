package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// Start server at port 8080!
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", dolarOperation)
	fmt.Println("Server started on PORT", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// Function that takes care of all the fetching and formatting
func dolarOperation(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	// Defines the interfaces which will contain the values retrieved with goquery
	nombre := []interface{}{}
	compra := []interface{}{}
	venta := []interface{}{}
	actualizado := []interface{}{}
	// URL that is gonna be scraped
	url := "https://www.cronista.com/MercadosOnline/dolar.html"

	// Make an HTTP GET request to the URL
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find in the document the table with the class "name", extract text and remove the "Dolar" part of the string
	doc.Find("td[class='name']").Each(func(i int, s *goquery.Selection) {
		value := s.Text()
		removeDolar := value[6:]
		nombre = append(nombre, removeDolar)
	})
	// Find in the document all the elements with the class "buy-value", extract the text and then use the function formatToNumber to make it a float
	doc.Find(".buy-value").Each(func(i int, s *goquery.Selection) {
		value := s.Text()
		parsed := formatToNumber(value)
		compra = append(compra, parsed)
	})

	// The previous function only returns 5 numbers but the others interfaces will have 6 elements so it is necessary to add a null, undefined or 0 in the position 2.
	compra = append(compra, nil)
	copy(compra[3:], compra[2:])
	compra[2] = nil

	// Find in the document all the element with the class "sell-value", extract text and the use the function formatToNumber on it
	doc.Find(".sell-value").Each(func(i int, s *goquery.Selection) {
		// Get the text within the element
		value := s.Text()
		parsed := formatToNumber(value)
		venta = append(venta, parsed)
	})

	// Find in the document all the element with the class "date" that is also in a table and then remove the "Actualizado" part in the string. Also replace the dots with "/"
	doc.Find("td[class='date']").Each(func(i int, s *goquery.Selection) {
		// Get the text within the element
		value := s.Text()
		formatToDate := strings.Replace(value, ".", "/", 3)
		removeActualizado := formatToDate[13:]
		actualizado = append(actualizado, removeActualizado)
	})

	// Create a struct to represent each object in the JSON array
	type JSONObject struct {
		Nombre      interface{} `json:"nombre"`
		Compra      interface{} `json:"compra"`
		Venta       interface{} `json:"venta"`
		Actualizado interface{} `json:"actualizado"`
	}

	// Populate the JSON array with objects containing elements from the arrays
	var jsonArray []JSONObject
	for i := 0; i < 6; i++ {
		jsonObject := JSONObject{
			Nombre:      nombre[i],
			Compra:      compra[i],
			Venta:       venta[i],
			Actualizado: actualizado[i],
		}
		jsonArray = append(jsonArray, jsonObject)
	}
	// Marshal the JSON array into a JSON string
	jsonData, err := json.Marshal(jsonArray)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header and write the JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// The function that takes care of removing the dolar sign, replace the comma for a dot so it can be parsed as a float64
func formatToNumber(str string) float64 {
	removeSign := strings.Replace(str, "$", "", 1)
	canConvert := strings.Replace(removeSign, ",", ".", 1)
	parsed, _ := strconv.ParseFloat(canConvert, 64)
	return parsed
}
