package meal_email

import (
	"fmt"
	"meals/meal_collection"
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type PDFGenerator interface {
	GenerateIngredientsPDF(ingredients []meal_collection.Ingredient) ([]byte, error)
}

type DefaultPDFGenerator struct{}

func (d DefaultPDFGenerator) GenerateIngredientsPDF(ingredients []meal_collection.Ingredient) ([]byte, error) {
	htmlContent := buildHTMLContent(ingredients)
	pdfBytes, err := convertHTMLToPDF(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("error converting HTML to PDF: %w", err)
	}
	return pdfBytes, nil
}

func contains(slice []meal_collection.Aisle, item meal_collection.Aisle) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func convertHTMLToPDF(html string) ([]byte, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %v", err)
	}

	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Cover.Zoom.Set(1.0)

	page := wkhtmltopdf.NewPageReader(strings.NewReader(html))
	pdfg.AddPage(page)

	if err := pdfg.Create(); err != nil {
		return nil, fmt.Errorf("failed to create PDF: %v", err)
	}

	return pdfg.Bytes(), nil
}

func buildHTMLContent(ingredients []meal_collection.Ingredient) string {
	const FONT_SIZE = 16
	const CHECKBOX_SIZE = FONT_SIZE - 2
	const LARGER_FONT_SIZE = 18
	const LARGER_CHECKBOX_SIZE = LARGER_FONT_SIZE - 2

	// Preallocate an estimated capacity for the builder.
	var sb strings.Builder
	sb.Grow(1024)

	// Write the header of the HTML document.
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Aisle Items</title>
	<style>
		html, body {
			margin: 0;
			padding: 0;
			width: 100%;
			height: 100%;
		}

		table {
			width: 100%;
			height: 50%;
			border-collapse: collapse;
			table-layout: fixed; /* Forces fixed column widths */
		}

		tr {
			height: 50%;
		}

		td {
			box-sizing: border-box;
			border: 1px solid #333;
			background-color: #ffffff;
			vertical-align: top;
			padding: 5px;
		}
		
		.cell-three {
			width: calc(100% / 3);
		}
		.cell-two {
			width: calc(100% / 2);
		}
		
		h3 {
			margin: 0 0 5px;
			padding: 2px;
			background-color: #00d5ff;
			text-align: center;
			font-size: 16px;
		}
		
		.checkbox-group label {
			font-size: ` + fmt.Sprintf("%d", FONT_SIZE) + `px;
			margin: 0;  /* Remove extra margin */
			padding: 0; /* Remove extra padding */
		}
		
		.cell-two .checkbox-group label {
			font-size: ` + fmt.Sprintf("%d", LARGER_FONT_SIZE) + `px;
		}
		
		input[type="checkbox"] {
			width: ` + fmt.Sprintf("%d", CHECKBOX_SIZE) + `px;
			height: ` + fmt.Sprintf("%d", CHECKBOX_SIZE) + `px;
			margin: 2px;  /* Add 2px margin around checkboxes */
			padding: 0;
		}
		
		.cell-two input[type="checkbox"] {
			width: ` + fmt.Sprintf("%d", LARGER_CHECKBOX_SIZE) + `px;
			height: ` + fmt.Sprintf("%d", LARGER_CHECKBOX_SIZE) + `px;
		}
		
		.page-break {
			page-break-after: always;
			break-after: page;
		}
	</style>
</head>
<body>
`)

	// Function to open a new table.
	openTable := func() {
		sb.WriteString("<table>\n")
	}
	// Function to close the current table.
	closeTable := func() {
		sb.WriteString("</table>\n")
	}

	rowStarters := []meal_collection.Aisle{meal_collection.AisleCheeseAndBakery, meal_collection.AisleFreezer, meal_collection.AisleBreakfastAndBaking, meal_collection.AisleProduce}
	rowClosers := []meal_collection.Aisle{meal_collection.AisleAlcoholButterCheese, meal_collection.AisleBeveragesAndSnacks, meal_collection.AislePastaGlobalCanned, meal_collection.AisleMeatAndYogurt}
	rowThree := []meal_collection.Aisle{meal_collection.AisleFreezer, meal_collection.AisleNoFoodItems, meal_collection.AisleBeveragesAndSnacks}
	// Generate table cells.
	for i, aisle := range meal_collection.AllAisles {
		// Start a new row if needed.
		if contains(rowStarters, aisle) {
			openTable()
			sb.WriteString("  <tr>\n")
		}

		var aisleHTML string
		if contains(rowThree, aisle) {
			aisleHTML = buildAisleCellHTML(aisle, ingredients, "cell-three")
		} else {
			aisleHTML = buildAisleCellHTML(aisle, ingredients, "cell-two")
		}
		sb.WriteString(aisleHTML)

		if contains(rowClosers, aisle) {
			sb.WriteString("  </tr>\n")
			closeTable()
		}

		if i == 4 {
			sb.WriteString(`<div class="page-break"></div>` + "\n")
		}
	}

	// Write the closing tags.
	sb.WriteString(`</body>
</html>`)

	return sb.String()
}

func buildAisleCellHTML(aisle meal_collection.Aisle, ingredients []meal_collection.Ingredient, cellClass string) string {
	longerColumns := []meal_collection.Aisle{meal_collection.AisleFreezer, meal_collection.AisleNoFoodItems, meal_collection.AisleBeveragesAndSnacks}

	var sb strings.Builder
	// Use the provided cellClass in the td element.
	sb.WriteString(fmt.Sprintf("    <td class=\"%s\">\n      <h3>%s</h3>\n", cellClass, aisle))

	// Filter ingredients for the current aisle.
	var itemsForAisle []meal_collection.Ingredient
	for _, ing := range ingredients {
		if ing.Aisle == aisle {
			itemsForAisle = append(itemsForAisle, ing)
		}
	}

	sb.WriteString("      <div class=\"checkbox-group\">\n")
	totalCheckboxes := 28
	if contains(longerColumns, aisle) {
		totalCheckboxes = 33
	}
	for i := 0; i < totalCheckboxes; i++ {
		if i < len(itemsForAisle) {
			ing := itemsForAisle[i]
			sb.WriteString(fmt.Sprintf("        <label><input type=\"checkbox\" disabled> %s</label><br>\n", ing.StringBolded()))
		} else {
			sb.WriteString("        <label><input type=\"checkbox\" disabled> </label><br>\n")
		}
	}
	sb.WriteString("      </div>\n")
	sb.WriteString("    </td>\n")
	return sb.String()
}
