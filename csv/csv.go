package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/commoddity/bank-informer/persistence"
)

const defaultFilename = "crypto_values.csv"

func WriteCryptoValuesToCSV(p *persistence.Persistence, cryptos []string) error {
	currentDate := time.Now().Format("2006-01-02")

	// Read existing records
	records, err := readCSV(defaultFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if the file has headers
	hasHeaders := len(records) > 0 && records[0][0] == "date"

	updated := false
	totalFiatBalance := 0.0
	for _, crypto := range cryptos {
		key := fmt.Sprintf("%s-%s", crypto, currentDate)
		avgValues, err := p.GetAverageCryptoValues(key)
		if err != nil {
			continue
		}

		record := []string{
			currentDate,
			crypto,
			fmt.Sprintf("%f", avgValues.CryptoBalance),
			fmt.Sprintf("%f", avgValues.FiatValue),
			fmt.Sprintf("%f", avgValues.FiatBalance),
		}

		totalFiatBalance += avgValues.FiatBalance

		_, records = updateOrAddRecord(records, record)
	}

	// Add total row
	totalRow := []string{
		currentDate,
		"TOTAL",
		"",
		"",
		fmt.Sprintf("%f", totalFiatBalance),
	}
	updated, records = updateOrAddRecord(records, totalRow)

	// Rewrite the CSV file only if updated
	if updated {
		return writeCSV(defaultFilename, records, !hasHeaders)
	}

	return nil
}

func readCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return [][]string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func updateOrAddRecord(records [][]string, newRecord []string) (bool, [][]string) {
	updated := false
	for i, record := range records {
		if record[0] == newRecord[0] && record[1] == newRecord[1] {
			records[i] = newRecord
			updated = true
			break
		}
	}
	if !updated {
		records = append(records, newRecord)
		updated = true
	}
	return updated, records
}

func writeCSV(filePath string, records [][]string, writeHeaders bool) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// Write headers if needed
	if writeHeaders {
		headers := []string{"date", "cryptoSymbol", "cryptoBalance", "fiatValue", "fiatBalance"}
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	return writer.WriteAll(records)
}
