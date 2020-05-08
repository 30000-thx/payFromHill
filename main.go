package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sclevine/agouti"
)

const (
	FIRST_PAGE = "https://e-banking2.hangseng.com/1/3/!ut/p/z1/jY1BDoIwFETP4gk6LS3IsqJAQSGKBeyGsBKiogujiacXMC5cGJ3d_28mjxhSEtPVt3ZfX9tzVx_7e2fsynUlCylnITwmIAOVRo5nA6CkGAtAvnRnENEUNo2TyB_gGGIGHFg60RTZHKtwq5wNZW9cfApSkUmseUIF28RA-tr_9HuBDLmz7D-5WkDxGYdY-RTK-m-PL5E__dmjqZoDuZy01uV98gR_3goJ/dz/d5/L2dJQSEvUUt3QS80TmxFL1o2Xzk5QTJIMTQySDBDMjUwQUdJT0o3QzYwMDAx/?cmd-All_="
	// URL_BILLS = "https://www.ppshk.com/pps/AppLoadBill"
	IRD_MerchantCode = "10"
)

var chromeDriverLocation string
var waitLogin chan PayAction
var currentPa *PayAction

type Step int

const (
	Step_SetConfig Step = iota
	Step_WaitLogin
	Step_Pay
)

type Status struct {
	CurrentStep Step `json:"currentStep"`
	IsLogin     bool `json:"isLogon"`
}

var status Status

//embeded DB for cache
var db *bolt.DB

func main() {
	//TODO: change to download from internet
	if runtime.GOOS == "windows" {
		log.Println("Hello from Windows")
		chromeDriverLocation = "selenium/chromedriver.exe"
	} else if runtime.GOOS == "darwin" {
		log.Println("Hello from Mac")
		chromeDriverLocation = "selenium/chromedriver"
	} else {
		log.Println("Hello from", runtime.GOOS, "not supporting...")
		return
	}
	waitLogin = make(chan PayAction)

	d, err := bolt.Open("autoTax.db", 0600, nil)
	db = d
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//go startChrome()
	go webserver()

	select {}
}

func setPayAction(_pa *PayAction) error {
	if status.CurrentStep != Step_SetConfig {
		return errors.New("Unknown Action")
	}

	pa = _pa
	go startChrome()

	go monitonLogon()

	status.CurrentStep = Step_WaitLogin

	return nil
}

var monitonLogonCh chan bool
var page *agouti.Page
var pa *PayAction

func startChrome() {

	command := []string{chromeDriverLocation, "--port={{.Port}}"}
	driver := agouti.NewWebDriver("http://{{.Address}}", command)

	if err := driver.Start(); err != nil {
		fmt.Println("Failed to start driver:", err)
		return
	}
	//defer driver.Stop()
	var err error

	page, err = driver.NewPage()
	if err != nil {
		fmt.Println("Failed to open page:", err)
		return
	}

	fmt.Println("navigating to", FIRST_PAGE)
	if err := page.Navigate(FIRST_PAGE); err != nil {
		fmt.Println("Failed to navigate:", err)
		return
	}

}

func startPay() {

	if status.CurrentStep != Step_WaitLogin {
		return //"Another thread is workining..."
	}
	status.CurrentStep = Step_Pay

	click(page, `document.querySelector("#serviceNavItem-2-2-2-1>a")`)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	for pa.PaidAmount <= pa.Total {
		fmt.Println("navigating to bill home")

		click(page, `document.querySelector("#serviceNavItem-2-2-1 a")`)
		wait()

		click(page, `document.querySelector("[id$='MerchantList']")`)
		wait()

		click(page, `document.evaluate('//td[text()="INLAND REVENUE DEPARTMENT"]', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue`)
		wait()

		putValue(page, `document.querySelector("[name$='_BillAcct']")`, pa.AccountNumber)
		wait()

		click(page, `document.querySelector("[id$='_BillTypeList']")`)
		wait()

		click(page, `document.evaluate('//td[text()="01 TAX DEMAND NOTE"]', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue`)
		wait()

		//TODO: select from DebitAcctList

		cents := math.Floor(((r1.Float64() * 0.5) + 1) * 100)
		dollars := cents / 100
		fmt.Print("submitting", dollars)

		putValue(page, `document.querySelector("[name$='_PaymentAmt']")`, dollars)
		wait()

		click(page, `document.querySelector("#okBtn")`)
		wait()

		//Final Sumbit
		if pa.IsDebug {
			click(page, `document.querySelector(".btn-set>.btn-gry-2>a")`)
			//btn-gry-2
		} else {
			click(page, `document.querySelector(".btn-set>.btn-grn-1>a")`)
		}
		wait()

		//logPay(int(cents), current)
		pa.AddAmount(int(cents))
	}

	pa.Done()

	return
}

func wait() {
	time.Sleep(2 * time.Second)
}

func monitonLogon() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				status.IsLogin = checkLogon()
				if status.IsLogin {
					return
				}
			case <-monitonLogonCh:
				return
			}
		}
	}()
}

func checkLogon() bool {
	if page == nil {
		return false
	}

	var result bool
	page.RunScript(`return document.querySelector('.hase-gt-nav-log-off') != null`, nil, &result)

	if result {
		page.RunScript(`return document.querySelector('.hase-gt-loading-dialog').offsetParent == null`, nil, &result)
	}

	return result
}

func putValue(page *agouti.Page, selector string, value interface{}) {
	page.RunScript(fmt.Sprint(selector, ".value = ", value), nil, nil)
}

func click(page *agouti.Page, selector string) {
	fmt.Println("clicking: ", selector, ", Script: ", fmt.Sprint(selector, ".click()"))
	page.RunScript(fmt.Sprint(selector, ".click()"), nil, nil)
}

type PayAction struct {
	StartDate time.Time
	ID        string
	IsEnd     bool
	IsDebug   bool

	AccountNumber string
	Total         int //in cents

	PayRecords []PayRecord `json:"-"`

	LastPayRecord *PayRecord
	PaidAmount    int
}

type PayRecord struct {
	PayAmount int
	DateTime  time.Time
}

func (p *PayAction) AddAmount(amount int) {
	pr := PayRecord{
		PayAmount: amount,
		DateTime:  time.Now(),
	}
	p.PayRecords = append(p.PayRecords, pr)
	p.PaidAmount += amount

	p.LastPayRecord = &pr

	p.save()
	sendMsgToWs(p)
}

func (p *PayAction) Done() {
	p.IsEnd = true
	p.save()

	sendMsgToWs(p)
}

func (p *PayAction) save() {
	go func() {
		db.Update(func(tx *bolt.Tx) error {
			bu, err := tx.CreateBucketIfNotExists([]byte("PayAction"))
			if err != nil {
				// k.logger.Log("method", "putToDB", "err", err)
				return err
			}
			b, err := json.Marshal(p)
			if err != nil {
				return err
			}

			err = bu.Put([]byte(p.ID), b)
			if err != nil {
				// k.logger.Log("method", "putToDB", "err", err)
			}
			return err
		})
	}()
}

func GetPayActions() (actions []PayAction, err error) {

	err = db.View(func(tx *bolt.Tx) error {
		tx.Bucket([]byte("PayAction")).ForEach(func(k, v []byte) error {
			var pa PayAction
			err = json.Unmarshal(v, &pa)
			if err != nil {
				return err
			}
			actions = append(actions, pa)
			return nil
		})

		return nil
	})
	return

}
