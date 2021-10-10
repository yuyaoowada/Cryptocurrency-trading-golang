package main

import (
	"buy-btc/bitflyer"
	"fmt"
	"math"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	// "github.com/aws/aws-sdk-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// 注文方法：Limit（指値）の買い注文
// 価格→現在価格の95%
// 数量→0.001BTC

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	apiKey, err := getParameter("buy-btc-apikey")
	if err != nil {
		return getErrorResponse(err.Error()), err
	}

	apiSecret, err := getParameter("buy-btc-apisecret")
	if err != nil {
		return getErrorResponse(err.Error()), err
	}

	ticker, err := bitflyer.GetTicker(bitflyer.Btcjpy)

	buyPrice := RoundDecimal(ticker.Ltp * 0.95)

	order := bitflyer.Order{
		ProductCode:     bitflyer.Btcjpy.String(),
		ChildOrderType:  bitflyer.Limit.String(),
		Side:            bitflyer.Buy.String(),
		Price:           buyPrice,
		Size:            0.001,
		MinuteToExpires: 4320, // 3days
		TimeInForce:     bitflyer.Gtc.String(),
	}

	orderRes, err := bitflyer.PlaceOrder(&order, apiKey, apiSecret)
	if err != nil {
		return getErrorResponse(err.Error()), err
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("res:%+v", orderRes),
		StatusCode: 200,
	}, nil
}

func RoundDecimal(num float64) float64 {
	return math.Round(num)
}

// System Managerからパラメータを取得する関数
func getParameter(key string) (string, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := ssm.New(sess, aws.NewConfig().WithRegion("ap-northeast-1"))

	params := &ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(true),
	}

	res, err := svc.GetParameter(params)
	if err != nil {
		return "", err
	}
	return *res.Parameter.Value, nil
}

func getErrorResponse(message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       message,
		StatusCode: 400,
	}
}

func main() {
	lambda.Start(handler)
}
