package main

import (
	"net/http"
	"net/url"
	"io/ioutil"
	"log"
	"flag"

	"encoding/json"
	"github.com/tok3n/tok3nchain"
	"fmt"
	"os"
	"bufio"

	"golang.org/x/crypto/ed25519"
	"math/rand"
	"encoding/base64"
	"bytes"
	"time"
	"crypto/sha256"

	"github.com/Pallinder/go-randomdata"
	"strings"
)

var endpoint1 string
var endpoint2 string
var coordinator_endpoint string
var appPrivateKey ed25519.PrivateKey
var appPublicKey ed25519.PublicKey


func main() {
	flag.StringVar(&endpoint1, "endpoint 1", "localhost:8081", "url of the endpoint 1")
	flag.StringVar(&endpoint2, "endpoint 2", "localhost:8080", "url of the endpoint 2")
	flag.StringVar(&coordinator_endpoint, "coordinator", "localhost:8080", "url of the coordinator endpoint")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Initin coorinator\n")
	reader.ReadString('\n')

	coordpubl, coordpriv, err := coordinator_init(coordinator_endpoint)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Adding node 1\n")
	reader.ReadString('\n')

	_, err = coordinator_addNode(coordpriv, endpoint1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Adding node 2\n")
	reader.ReadString('\n')

	_, err = coordinator_addNode(coordpriv, endpoint2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Creating App\n")
	reader.ReadString('\n')

	//transapp, err, app_public, app_private := createAPP()
	transapp, err, _, _ := createAPP(coordinator_endpoint,coordpubl, coordpriv)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("App created: %+v\n\n\n",transapp)





	for n := 0; n<20; n++{
		fmt.Printf("Creating User\n")
		reader.ReadString('\n')

		//newuser, err, user_public, user_private := createUser(endpoint1)
		newuser, err, _, _ := createUser(endpoint1)
		if err != nil{
			log.Fatal(err)
		}

		fmt.Printf("User created: %+v\n\n\n",newuser)
	}

/*
	contract, err := createContract(app_public, app_private, user_public, user_private)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("Contract created: %+v\n",contract)
	fmt.Print("Next ")
	reader.ReadString('\n')

	signreq_created_1, err := create_SignRequest("UserAsk",contract)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("SignRequest created: %+v\n",signreq_created_1)
	fmt.Print("Next ")
	reader.ReadString('\n')


	signreq_fetched_1, err := get_SignRequest(signreq_created_1.IdVal)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("SignRequest fetched: %+v\n",signreq_fetched_1)
	fmt.Print("Next ")
	reader.ReadString('\n')

	contract01, err := signAndSend_SignRequest(signreq_fetched_1, user_public, user_private)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("Sign contract: %+v\n",contract01)
	fmt.Print("Next ")
	reader.ReadString('\n')















	signreq_created_2, err := create_SignRequest("UserValidate",contract)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("SignRequest created: %+v\n",signreq_created_2)
	fmt.Print("Next ")
	reader.ReadString('\n')


	signreq_fetched_2, err := get_SignRequest(signreq_created_2.IdVal)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("SignRequest fetched: %+v\n",signreq_fetched_2)
	fmt.Print("Next ")
	reader.ReadString('\n')

	contract02, err := signAndSend_SignRequest(signreq_fetched_2, user_public, user_private)
	if err != nil{
		log.Fatal(err)
	}

	fmt.Printf("Sign contract: %+v\n",contract02)
	fmt.Print("Next ")
	reader.ReadString('\n')




*/


}

func coordinator_init(coordinator string)([]byte, []byte, error){
	appPublicKey, appPrivateKey, err := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
	if err != nil{
		return nil, nil, err
	}

	base64content := base64.StdEncoding.EncodeToString(appPublicKey)
	returned, err := getRemote("http://"+coordinator+"/api/v1/coord/register_masterkey?url="+coordinator_endpoint+"&key="+url.QueryEscape( base64content) )
	if err != nil{
		return nil, nil, err
	}

	if returnedstr := string(returned); returnedstr != "OK" {
		return nil, nil, fmt.Errorf(returnedstr)
	}

	return appPublicKey, appPrivateKey, nil

}

func coordinator_addNode(coordinator_private []byte, node string)(string, error){
	//ARRAY OF URLS
	data := struct {
		Urls []string
	}{
		[]string{node},
	}

	jsonstr, err := json.Marshal(data)
	if err != nil{
		return "", err
	}

	sha_256 := sha256.New()
	sha_256.Write(jsonstr)
	contentsha :=  sha_256.Sum(nil)
	base64content := base64.StdEncoding.EncodeToString(jsonstr)

	sign := ed25519.Sign(coordinator_private, contentsha)
	base64sign := base64.StdEncoding.EncodeToString(sign)

	tosend := struct{
		Content string
		Sign	string
	}{
		base64content,
		base64sign,
	}

	jsonstr, err = json.Marshal(tosend)
	if err != nil{
		return "", err
	}
	response, err := postRemote("http://"+coordinator_endpoint+"/api/v1/coord/register_nodes",jsonstr)
	if err != nil {
		return "", err
	}

	if responsestr := string(response); strings.HasPrefix(responsestr,"ERROR"){
		return "", fmt.Errorf(responsestr)
	}else{
		return responsestr, nil
	}

}

func signAndSend_SignRequest(transaction *tok3nchain.Transaction_Serializable, user_public,user_private []byte, endpoint string)(*tok3nchain.Transaction_Serializable,error){
	fmt.Printf("Sign Request\n")
	jsonstr, err := json.Marshal(transaction)
	if err != nil{
		return nil,err
	}
	transaction.Content = base64.StdEncoding.EncodeToString(jsonstr)

	sha_256 := sha256.New()
	sha_256.Write(jsonstr)
	contentsha :=  sha_256.Sum(nil)
	transaction.Hash = base64.StdEncoding.EncodeToString(contentsha)

	sign := ed25519.Sign(user_private, contentsha)
	transaction.Sign = base64.StdEncoding.EncodeToString(sign)

	transaction.Signer = base64.StdEncoding.EncodeToString(user_public)

	jsonstr, err = json.Marshal(transaction)
	if err != nil{
		return nil,err
	}

	response, err := postRemote("http://"+endpoint+"/api/v1/sign_contract",jsonstr)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response,transaction)
	if err != nil {
		fmt.Printf("%s\n",string(response))
	}

	return transaction, err
}

func get_SignRequest(id int64, endpoint string)(*tok3nchain.Transaction_Serializable,error){
	fmt.Printf("Get Sign Request\n")
	response, err := getRemote(fmt.Sprintf("http://%s/api/v1/get_signingRequest?id=%d",endpoint,id))
	if err != nil {
		log.Printf("return: %s",string(response))
		return nil, err
	}
	transaction := new(tok3nchain.Transaction_Serializable)
	err = json.Unmarshal(response,transaction)
	if err != nil {
		fmt.Printf("%s\n",string(response))
	}

	return transaction, err
}

func create_SignRequest(signKind string, parent *tok3nchain.Transaction_Serializable, endpoint string) (*tok3nchain.Transaction_Serializable,error) {
	fmt.Printf("Creating transaction sign request\n")


	transaction := new(tok3nchain.Transaction_Serializable)
	transaction.Payload = parent.Payload
	transaction.Parent = parent.IdVal
	transaction.ParentBlock = parent.BlockId
	transaction.AppID = parent.AppID
	transaction.SignerKinds = parent.SignerKinds
	transaction.SignKind = signKind
	transaction.Callback = parent.Callback



	jsonstr, err := json.Marshal(transaction)
	if err != nil{
		return nil,err
	}

	response, err := postRemote("http://"+endpoint+"/api/v1/create_signingRequest",jsonstr)
	if err != nil {
		log.Printf("return: %s",string(response))
		return nil, err
	}

	err = json.Unmarshal(response,transaction)
	if err != nil {
		fmt.Printf("%s\n",string(response))
	}

	return transaction, err
}


func createContract(app_public,app_private,user_public,user_private []byte, endpoint string) (*tok3nchain.Transaction_Serializable,error) {
	fmt.Printf("Creating Contract creation\n")
	block, err := getBlock(endpoint)
	if err != nil {
		return nil, err
	}

	appPublicKey, _ := app_public, app_private

	transaction := new(tok3nchain.Transaction_Serializable)
	transaction.SignerKinds = []string{"UserAsk","UserValidate"}
	transaction.SignKind = "__NEWCONTRACT"
	transaction.AppID = base64.StdEncoding.EncodeToString(appPublicKey)
	transaction.Parent = 0
	transaction.Callback = "http://localhost:8081/sign_data"
	transaction.Payload = randomdata.Email()

	transaction.BlockId = block.IdVal
	transaction.Block = block.Hash
	transaction.Creation = time.Now().UnixNano()


	jsonstr, err := json.Marshal(transaction)
	if err != nil{
		return nil,err
	}
	transaction.Content = base64.StdEncoding.EncodeToString(jsonstr)

	sha_256 := sha256.New()
	sha_256.Write(jsonstr)
	contentsha :=  sha_256.Sum(nil)
	transaction.Hash = base64.StdEncoding.EncodeToString(contentsha)

	sign := ed25519.Sign(user_private, contentsha)
	transaction.Sign = base64.StdEncoding.EncodeToString(sign)

	transaction.Signer = base64.StdEncoding.EncodeToString(user_public)

	jsonstr, err = json.Marshal(transaction)
	if err != nil{
		return nil,err
	}

	response, err := postRemote("http://"+endpoint+"/api/v1/create_contract",jsonstr)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(response,transaction)
	if err != nil {
		fmt.Printf("%s\n",string(response))
	}

	return transaction, err
}

func createUser(endpoint string) (*tok3nchain.Transaction_Serializable,error,[]byte,[]byte) {
	fmt.Printf("Creating new User\n")
	block, err := getBlock(endpoint)
	if err != nil {
		return nil, err, nil, nil
	}

	appPublicKey, appPrivateKey, err = ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
	if err != nil{
		return nil, err, nil, nil
	}

	sha_256 := sha256.New()


	transaction := new(tok3nchain.Transaction_Serializable)
	transaction.SignerKinds = []string{"NewUser"}
	transaction.SignKind = "NewUser"
	transaction.AppID = base64.StdEncoding.EncodeToString(appPublicKey)
	transaction.Parent = 0
	transaction.Callback = "http://localhost:8081"
	transaction.Payload = randomdata.Email()

	transaction.BlockId = block.IdVal
	transaction.Block = block.Hash
	transaction.Creation = time.Now().UnixNano()


	jsonstr, err := json.Marshal(transaction)
	if err != nil{
		return nil,err, nil, nil
	}
	transaction.Content = base64.StdEncoding.EncodeToString(jsonstr)

	sha_256.Write(jsonstr)
	contentsha :=  sha_256.Sum(nil)
	transaction.Hash = base64.StdEncoding.EncodeToString(contentsha)

	sign := ed25519.Sign(appPrivateKey, contentsha)
	transaction.Sign = base64.StdEncoding.EncodeToString(sign)

	transaction.Signer = transaction.AppID

	jsonstr, err = json.Marshal(transaction)
	if err != nil{
		return nil,err, nil, nil
	}

	response, err := postRemote("http://"+endpoint+"/api/v1/post_single_transaction",jsonstr)
	if err != nil {
		return nil, err, nil, nil
	}

	err = json.Unmarshal(response,transaction)

	return transaction, err, appPublicKey, appPrivateKey
}

func createAPP(endpoint string, coord_publ, coord_priv []byte) (*tok3nchain.Transaction_Serializable,error, []byte, []byte) {
	fmt.Printf("Creating new App\n")
	block, err := getBlock(endpoint)
	if err != nil {
		return nil, err, nil, nil
	}

	appPublicKey, appPrivateKey, err = ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().UnixNano())))
	if err != nil{
		return nil, err, nil, nil
	}

	sha_256 := sha256.New()


	transaction := new(tok3nchain.Transaction_Serializable)
	transaction.SignerKinds = []string{"NewApp"}
	transaction.SignKind = "NewApp"
	transaction.AppID = base64.StdEncoding.EncodeToString(appPublicKey)
	transaction.Parent = 0
	transaction.Callback = "http://localhost:8081"
	transaction.Payload = "Test app1"
	transaction.OriginatorURl = ""

	transaction.BlockId = block.IdVal
	transaction.Block = block.Hash
	transaction.Creation = time.Now().UnixNano()


	jsonstr, err := json.Marshal(transaction)
	if err != nil{
		return nil,err, nil, nil
	}
	transaction.Content = base64.StdEncoding.EncodeToString(jsonstr)

	sha_256.Write(jsonstr)
	contentsha :=  sha_256.Sum(nil)
	transaction.Hash = base64.StdEncoding.EncodeToString(contentsha)

	sign := ed25519.Sign(coord_priv, contentsha)
	transaction.Sign = base64.StdEncoding.EncodeToString(sign)

	transaction.Signer = base64.StdEncoding.EncodeToString(coord_publ)

	jsonstr, err = json.Marshal(transaction)
	if err != nil{
		return nil,err, nil, nil
	}

	response, err := postRemote("http://"+endpoint+"/api/v1/coord/add_app",jsonstr)
	if err != nil {
		return nil, err, nil, nil
	}

	if (strings.HasPrefix(string(response),"ERROR")){
		return nil, fmt.Errorf("Error form server: %s",string(response)),nil,nil
	}

	return transaction, err, appPublicKey, appPrivateKey
}


func getBlock(endpoint string) (*tok3nchain.Block_Serializable,error) {
	response, err := callRemote("http://"+endpoint+"/api/v1/last_block")

	if err != nil {
		return nil, err
	}

	block := new(tok3nchain.Block_Serializable)
	err = json.Unmarshal(response,block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func callRemote(url string)([]byte, error){
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func postRemote(url string, data []byte)([]byte, error){
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return body, nil
}

func getRemote(url string)([]byte, error){
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return body, nil
}