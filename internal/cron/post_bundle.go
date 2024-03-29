package cron

import (
	"log"

	"github.com/everFinance/goar/types"
	"github.com/liteseed/argo/transaction"
	"github.com/liteseed/edge/internal/database/schema"
)

func parseDataItemFromOrder(c *Context, o *schema.Order) (*transaction.DataItem, error) {
	rawDataItem, err := c.store.Get(o.StoreID)
	if err != nil {
		return nil, err
	}
	dataItem, err := transaction.DecodeDataItem(rawDataItem)
	if err != nil {
		return nil, err
	}
	err = c.database.UpdateStatus(o.ID, schema.Sent)
	if err != nil {
		return nil, err
	}
	return dataItem, nil
}

func (c *Context) postBundle() {
	o, err := c.database.GetQueuedOrders(25)
	if err != nil {
		return
	}

	if len(*o) == 0 {
		log.Println("no dataitem to post")
		return
	}

	dataItems := []transaction.DataItem{}

	for _, order := range *o {
		dataItem, err := parseDataItemFromOrder(c, &order)
		if err != nil {
			log.Println(err)
			log.Println("failed to decode:", order.StoreID)
			continue
		}
		dataItems = append(dataItems, *dataItem)
	}

	bundle, err := transaction.NewBundle(&dataItems)
	if err != nil {
		log.Println("failed to bundle:", err)
		return
	}

	tx, err := c.wallet.SendData([]byte(bundle.RawData), []types.Tag{{Name: "Bundle-Format", Value: "binary"}, {Name: "Bundle-Version", Value: "2.0.0"}})
	if err != nil {
		log.Println("failed to upload:", err)
		return
	}
	log.Println(tx.ID)
}
