package main

func main() {
	if err := generateAWSFlatBuffers(); err != nil {
		panic(err)
	}
	if err := generateGCPFlatBuffers(); err != nil {
		panic(err)
	}
	if err := generateAzureFlatBuffers(); err != nil {
		panic(err)
	}
}
