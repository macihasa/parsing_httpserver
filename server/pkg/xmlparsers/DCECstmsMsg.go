package xmlparsers

import (
	"encoding/csv"
	"encoding/xml"
	"log"
	"os"
	"sync"
)

type CstmsInterface struct {
	XMLName xml.Name `xml:"CstmsInterface"`
	Shps    []Shp    `xml:"Dtls>Entries>Entry>Mvmts>Mvmt>TDOCs>TDOC>Shps>Shp"`
}

type Shp struct {
	HAWB          string       `xml:"HAWB"`
	LineItems     []LineItem   `xml:"LineItems>LineItem"`
	GrossWgt      string       `xml:"GrossWgt"`
	ActWgt        string       `xml:"ActWgt"`
	Cube          string       `xml:"Cube"`
	Incoterms     string       `xml:"Incoterms"`
	DHLServiceCd  string       `xml:"DHLServiceCd"`
	CargoDesc     string       `xml:"CargoDesc"`
	BillProdCd    string       `xml:"BillProdCd"`
	ProdContentCd string       `xml:"ProdContentCd"`
	LclLangCd     string       `xml:"LclLangCd"`
	PackagingType string       `xml:"PackagingType"`
	TotPackages   string       `xml:"TotPackages"`
	TxAndDty      TxAndDty     `xml:"TxAndDty"`
	ReasonForExpt string       `xml:"ReasonForExpt"`
	ShpPayer      ShpPayer     `xml:"ShpPayer"`
	CnsgnrCoDtls  CnsgnrCoDtls `xml:"ShpCnsgnr>CnsgnrCoDtls"`
	CnsgneCoDtls  CnsgneCoDtls `xml:"ShpCnsgne>CnsgneCoDtls"`
	ExpterCoDtls  ExpterCoDtls `xml:"ShpExpter>ExpterCoDtls"`
	ImpterCoDtls  ImpterCoDtls `xml:"ShpImpter>ImpterCoDtls"`
	DatElGrp      []DatEl      `xml:"DatElGrp>DatEl"`
}

type LineItem struct {
	GoodsItemNo          string `xml:"GoodsItemNo"`
	TariffQnty           string `xml:"TariffQnty"`
	MeasureUnitQualifier string `xml:"MeasureUnitQualifier"`
	DescOfGoods          string `xml:"DescOfGoods"`
	TariffCdNo           string `xml:"TariffCdNo"`
	CtryMfctrerOrgn      string `xml:"CtryMfctrerOrgn"`
	InvNo                string `xml:"InvNo"`
	InvLineVal           string `xml:"InvLineVal"`
	InvoiceDate          string `xml:"InvoiceDate"`
	InvCrncyCd           string `xml:"InvCrncyCd"`
	NetWeight            string `xml:"NetWeight"`
	GrossWgt             string `xml:"GrossWgt"`
	UnitPrice            string `xml:"UnitPrice"`
	CtryOrgnCd           string `xml:"CtryOrgnCd"`
	InvoiceLineNo        string `xml:"InvoiceLineNo"`
}

type TxAndDty struct {
	CstmsVal        string `xml:"CstmsVal"`
	CstmsValCrncyCd string `xml:"CstmsValCrncyCd"`
}

type ShpPayer struct {
	DHLAcctNo string `xml:"DHLAcctNo"`
}

type CnsgnrCoDtls struct {
	CoName   string  `xml:"CoName"`
	TraderID string  `xml:"TraderID"`
	AddrEng  AddrEng `xml:"AddrEng"`
}

type CnsgneCoDtls struct {
	CoName   string  `xml:"CoName"`
	TraderID string  `xml:"TraderID"`
	AddrEng  AddrEng `xml:"AddrEng"`
}

type ExpterCoDtls struct {
	CoName   string  `xml:"CoName"`
	TraderID string  `xml:"TraderID"`
	AddrEng  AddrEng `xml:"AddrEng"`
}

type ImpterCoDtls struct {
	CoName   string  `xml:"CoName"`
	TraderID string  `xml:"TraderID"`
	AddrEng  AddrEng `xml:"AddrEng"`
}

type AddrEng struct {
	AddrLn1    string `xml:"AddrLn1"`
	AddrLn2    string `xml:"AddrLn2"`
	AddrLn3    string `xml:"AddrLn3"`
	StreetName string `xml:"StreetName"`
	City       string `xml:"City"`
	PostalCd   string `xml:"PostalCd"`
	CtryCd     string `xml:"CtryCd"`
}

type DatEl struct {
	Cd  string `xml:"Cd"`
	Val string `xml:"Val"`
}

type safeMaps struct {
	mu      sync.Mutex
	hawbMap map[string]bool
}

// Limits the amount of reader routines that's active at one time.
const MAX_NUM_GOROUTINES = 16

// Main program iterating the folder "root" and sending off the routines
func DCECustomsMsg(fileName string, msgch <-chan []byte, finished chan<- bool) {

	var readwg = new(sync.WaitGroup)
	var writewg = new(sync.WaitGroup)

	var rowsch = make(chan []string, 1024)

	var maps = &safeMaps{hawbMap: make(map[string]bool)}

	writewg.Add(1)
	go writer(fileName, rowsch, writewg)

	for i := 0; i < MAX_NUM_GOROUTINES; i++ {
		readwg.Add(1)
		go reader(msgch, rowsch, maps, readwg)
	}

	// Wait until readers have finished writing all data to the rows channel before closing it.
	readwg.Wait()
	close(rowsch)

	// Wait for the writer to finish writing all data from the rows channel
	writewg.Wait()
	finished <- true
}

// Writer is run as single Goroutine that creates a csv file and writes the data from the rows channel onto it
func writer(fileName string, rowsch <-chan []string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Create(fileName + ".csv")
	if err != nil {
		log.Fatal("Unable to create output file: ", err)
	}

	csvWriter := csv.NewWriter(file)
	csvWriter.Comma = rune(';')
	writeHeaderRow(csvWriter)
	log.Println("Wrote header row to csv file")

	var rowindex int
	for row := range rowsch {
		err := csvWriter.Write(row)
		if err != nil {
			log.Println(err)
		}
		rowindex++
	}
	csvWriter.Flush()
	log.Println("Flushed csv writer")
	file.Close()
}

// reader is run as several Goroutines opening and parsing the contents of XML files to the rows channel
func reader(msgch <-chan []byte, rowsch chan<- []string, maps *safeMaps, wg *sync.WaitGroup) {
	defer wg.Done()
	for msg := range msgch {
		cstmsInterface := new(CstmsInterface)
		err := xml.Unmarshal(msg, cstmsInterface)
		if err != nil {
			log.Println("Unable to unmarshal msg ", err)
		}

		for _, shp := range cstmsInterface.Shps {
			maps.mu.Lock()
			_, ok := maps.hawbMap[shp.HAWB]
			if ok {
				maps.mu.Unlock()
				continue
			}
			maps.hawbMap[shp.HAWB] = true
			maps.mu.Unlock()
			for _, lineItem := range shp.LineItems {
				var transType string
				for _, datEl := range shp.DatElGrp {
					if datEl.Cd == "Clss" {
						transType = datEl.Val
					}
				}
					rowsch <- []string{
					shp.HAWB,
					shp.GrossWgt,
					shp.ActWgt,
					shp.CargoDesc,
					shp.DHLServiceCd,
					shp.Incoterms,
					transType,
					shp.ShpPayer.DHLAcctNo,
					shp.CnsgnrCoDtls.CoName,
					shp.CnsgnrCoDtls.AddrEng.AddrLn1,
					shp.CnsgnrCoDtls.AddrEng.City,
					shp.CnsgnrCoDtls.AddrEng.PostalCd,
					shp.CnsgnrCoDtls.AddrEng.CtryCd,
					shp.CnsgnrCoDtls.TraderID,
					shp.CnsgneCoDtls.CoName,
					shp.CnsgneCoDtls.AddrEng.AddrLn1,
					shp.CnsgneCoDtls.AddrEng.City,
					shp.CnsgneCoDtls.AddrEng.PostalCd,
					shp.CnsgneCoDtls.AddrEng.CtryCd,
					shp.CnsgneCoDtls.TraderID,
					shp.ExpterCoDtls.CoName,
					shp.ExpterCoDtls.AddrEng.AddrLn1,
					shp.ExpterCoDtls.AddrEng.City,
					shp.ExpterCoDtls.AddrEng.PostalCd,
					shp.ExpterCoDtls.AddrEng.CtryCd,
					shp.ExpterCoDtls.TraderID,
					shp.ImpterCoDtls.CoName,
					shp.ImpterCoDtls.AddrEng.AddrLn1,
					shp.ImpterCoDtls.AddrEng.City,
					shp.ImpterCoDtls.AddrEng.PostalCd,
					shp.ImpterCoDtls.AddrEng.CtryCd,
					shp.ImpterCoDtls.TraderID,
					lineItem.DescOfGoods,
					lineItem.TariffCdNo,
					lineItem.TariffQnty,
					lineItem.MeasureUnitQualifier,
					lineItem.InvLineVal,
					lineItem.InvCrncyCd,
					lineItem.NetWeight,
					lineItem.GrossWgt,
					lineItem.UnitPrice,
					lineItem.CtryOrgnCd,
					lineItem.InvoiceLineNo,
					lineItem.InvNo,
					lineItem.InvoiceDate}
			}
		}
	}
}

func writeHeaderRow(writer *csv.Writer) {
	writer.Write([]string{
		"HAWB",
		"GrossWgt",
		"ActWgt",
		"CargoDesc",
		"DHLServiceCd",
		"Incoterm",
		"TransType",
		"PayerDHLAcctNo",
		"Shipper Company Name",
		"Shipper Address Line 1",
		"Shipper City",
		"Shipper Postal Code",
		"Shipper Country Code",
		"Shipper TraderID",
		"Consignee Company Name",
		"Consignee Address Line 1",
		"Consignee City",
		"Consignee Postal Code",
		"Consignee Country Code",
		"Consignee TraderID",
		"Exporter Company Name",
		"Exporter Address Line 1",
		"Exporter City",
		"Exporter Postal Code",
		"Exporter Country Code",
		"Exporter TraderID",
		"Importer Company Name",
		"Importer Address Line 1",
		"Importer City",
		"Importer Postal Code",
		"Importer Country Code",
		"Importer TraderID",
		"Description of Goods",
		"Tariff Code Number",
		"Tariff Quantity",
		"Measurement Unit Qualifier",
		"Invoice Line Value",
		"Invoice Currency Code",
		"Net Weight",
		"Gross Weight",
		"Unit Price",
		"Country of Origin Code",
		"Invoice Line Number",
		"Invoice Number",
		"Invoice Date",
	})
}
