package main

import (
	"fmt"
	"github.com/gs2io/gs2-golang-sdk/account"
	"github.com/gs2io/gs2-golang-sdk/core"
	"github.com/gs2io/gs2-golang-sdk/inventory"
	"github.com/gs2io/gs2-golang-sdk/lottery"
	"github.com/openlyinc/pointy"
	"github.com/tidwall/gjson"
	"os"
)

// GS2の接続に必要なセッションを作成する
func loginGS2() core.Gs2RestSession {

	var session = core.Gs2RestSession{
		Credential: &core.BasicGs2Credential{
			ClientId:     "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			ClientSecret: "yyyyyyyyyyyyyyyyyyyyyyyyyyy",
		},
		Region: core.ApNortheast1,
	}
	if err := session.Connect(); err != nil {
		panic("error occurred")
	}
	return session
}

// ユーザIDを作成する
func createAccount(client account.Gs2AccountRestClient) string {

	result, err := client.CreateAccount(&account.CreateAccountRequest{
		NamespaceName: pointy.String("gacha_ns"),
	})

	if err != nil {
		panic("error occurred")
	}
	item := result.Item

	fmt.Printf("Create User ID: %s\n", *item.UserId)
	writeUserId(*item.UserId)

	return *item.UserId
}

// ユーザIDを取得する
func getUserId(session core.Gs2RestSession) string {

	client := account.Gs2AccountRestClient{
		Session: &session,
	}

	var userId = readUserId()
	if userId == "" {
		userId = createAccount(client)
	}
	return userId
}

// ユーザIDをファイルから読み込む
func readUserId() string {

	fileName := "userId.txt"
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return ""
	}

	return string(bytes)
}

// ユーザIDをファイルに書き込む
func writeUserId(userId string) {

	fileName := "userId.txt"
	err := os.WriteFile(fileName, []byte(userId), 0666)
	if err != nil {
		panic("file write error")
	}
}

// ガチャを引く
func drawByUserId(session core.Gs2RestSession, userId string) {

	// クライアントを取得する
	client := lottery.Gs2LotteryRestClient{
		Session: &session,
	}

	// 単発ガチャを引く
	var result, err = client.DrawByUserId(
		&lottery.DrawByUserIdRequest{
			NamespaceName: pointy.String("gacha_ns"),
			LotteryName:   pointy.String("gacha_sushi_event_001"),
			UserId:        pointy.String(userId),
			Count:         pointy.Int32(1),
			Config:        nil,
		},
	)
	if err != nil {
		panic("error occurred")
	}
	items := result.Items

	fmt.Printf("ガチャで出た寿司\n")

	for _, v := range items {
		for _, v2 := range v.AcquireActions {

			json := *v2.Request
			loopcnt := gjson.Get(json, "acquireCounts.#")

			for i := 0; i < int(loopcnt.Int()); i++ {
				itemName := gjson.Get(json, fmt.Sprintf("acquireCounts.%d.itemName", i))
				itemCount := gjson.Get(json, fmt.Sprintf("acquireCounts.%d.count", i))
				fmt.Printf("%s : %s個\n", itemName, itemCount)
			}
		}
	}
}

// インベントリ情報を表示する
func displayInventoryInfo(session core.Gs2RestSession, id string) {

	client := inventory.Gs2InventoryRestClient{
		Session: &session,
	}

	{
		var result, err = client.DescribeSimpleItemModels(
			&inventory.DescribeSimpleItemModelsRequest{
				NamespaceName: pointy.String("gacha_ns"),
				InventoryName: pointy.String("sushi"),
			},
		)

		if err != nil {
			panic("error occurred")
		}

		fmt.Printf("現在の寿司所有数\n")
		for _, v := range result.Items {

			var result, err = client.GetSimpleItemByUserId(
				&inventory.GetSimpleItemByUserIdRequest{
					NamespaceName: pointy.String("gacha_ns"),
					InventoryName: pointy.String("sushi"),
					UserId:        pointy.String(id),
					ItemName:      pointy.String(*v.Name),
				},
			)
			if err != nil {
				panic("error occurred")
			}
			item := result.Item
			fmt.Printf("%s : %d\n", *item.ItemName, *item.Count)
		}
	}
}

func main() {

	// GS2にログインする
	session := loginGS2()
	userId := getUserId(session)

	// ガチャを引く
	drawByUserId(session, userId)

	// インベントリ情報を表示する
	displayInventoryInfo(session, userId)

	session.Disconnect()
}
