package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/IBM-Cloud/bluemix-go/session"

	"github.com/IBM-Cloud/bluemix-go/api/account/accountv2"
	v1 "github.com/IBM-Cloud/bluemix-go/api/container/containerv1"
	"github.com/IBM-Cloud/bluemix-go/api/mccp/mccpv2"
	"github.com/IBM-Cloud/bluemix-go/trace"
)

func main() {

	var org string
	flag.StringVar(&org, "org", "", "Bluemix Organization")

	var space string
	flag.StringVar(&space, "space", "", "Bluemix Space")

	var zone string
	flag.StringVar(&zone, "zone", "", "Zone")

	var privateVlan string
	flag.StringVar(&privateVlan, "privateVlan", "", "Private Vlan")

	var publicVlan string
	flag.StringVar(&publicVlan, "publicVlan", "", "Public vlan")

	var updatePrivateVlan string
	flag.StringVar(&updatePrivateVlan, "updatePrivateVlan", "", "Private vlan")

	var updatePublicVlan string
	flag.StringVar(&updatePublicVlan, "updatePublicVlan", "", "Public vlan")

	var skipDeletion bool
	flag.BoolVar(&skipDeletion, "no-delete", false, "If provided will delete the resources created")

	flag.Parse()

	trace.Logger = trace.NewLogger("true")
	if org == "" || space == "" || privateVlan == "" || publicVlan == "" || updatePrivateVlan == "" || updatePublicVlan == "" || zone == "" {
		flag.Usage()
		os.Exit(1)
	}

	var clusterInfo = v1.ClusterCreateRequest{
		Name:        "my_cluster",
		Datacenter:  "dal10",
		MachineType: "u2c.2x4",
		WorkerNum:   1,
		PrivateVlan: privateVlan,
		PublicVlan:  publicVlan,
		Isolation:   "public",
	}

	sess, err := session.New()
	if err != nil {
		log.Fatal(err)
	}

	client, err := mccpv2.New(sess)

	if err != nil {
		log.Fatal(err)
	}

	region := sess.Config.Region
	orgAPI := client.Organizations()
	myorg, err := orgAPI.FindByName(org, region)

	if err != nil {
		log.Fatal(err)
	}

	spaceAPI := client.Spaces()
	myspace, err := spaceAPI.FindByNameInOrg(myorg.GUID, space, region)

	if err != nil {
		log.Fatal(err)
	}

	accClient, err := accountv2.New(sess)
	if err != nil {
		log.Fatal(err)
	}
	accountAPI := accClient.Accounts()
	myAccount, err := accountAPI.FindByOrg(myorg.GUID, region)
	if err != nil {
		log.Fatal(err)
	}

	target := v1.ClusterTargetHeader{
		OrgID:     myorg.GUID,
		SpaceID:   myspace.GUID,
		AccountID: myAccount.GUID,
	}

	clusterClient, err := v1.New(sess)
	if err != nil {
		log.Fatal(err)
	}
	clustersAPI := clusterClient.Clusters()

	out, err := clustersAPI.Create(clusterInfo, target)
	if err != nil {
		log.Fatal(err)
	}

	workerPoolAPI := clusterClient.WorkerPools()
	workerPoolRequest := v1.WorkerPoolRequest{
		WorkerPoolConfig: v1.WorkerPoolConfig{
			Name:        "test-workerpool",
			Size:        2,
			MachineType: "u2c.2x4",
			Isolation:   "public",
		},
		DiskEncryption: true,
	}
	resp, err := workerPoolAPI.CreateWorkerPool(out.ID, workerPoolRequest)
	if err != nil {
		log.Fatal(err)
	}
	workerPoolZone := v1.WorkerPoolZone{
		ID: zone,
		WorkerPoolZoneNetwork: v1.WorkerPoolZoneNetwork{
			PrivateVLAN: privateVlan,
			PublicVLAN:  publicVlan,
		},
	}
	err = workerPoolAPI.AddZone(out.ID, resp.ID, workerPoolZone)
	if err != nil {
		log.Fatal(err)
	}
	err = workerPoolAPI.UpdateZoneNetwork(out.ID, zone, resp.ID, updatePrivateVlan, updatePublicVlan)
	if err != nil {
		log.Fatal(err)
	}

	err = workerPoolAPI.ResizeWorkerPool(out.ID, resp.ID, 3)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := workerPoolAPI.GetWorkerPool(out.ID, resp.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Pool id is ", pool.ID)

	err = workerPoolAPI.RemoveZone(out.ID, zone, resp.ID)
	if err != nil {
		log.Fatal(err)
	}

	err = workerPoolAPI.DeleteWorkerPool(out.ID, resp.ID)
	if err != nil {
		log.Fatal(err)
	}

}
