provider "brightbox" {
	apiurl = "https://api.gb1s.brightbox.com"
	username = "neil@brightbox.co.uk"
	account = "acc-yw7i0"
}

#resource "brightbox_cloudip" "default" {
#	target = "${brightbox_server.first_server.id}"
#}

resource "brightbox_server" "first_server" {
	image = "img-m8vud"
	name = "Test terraform server"
	count = 2
	provisioner "remote-exec" {
	  inline = [
	    "sudo apt-get -y update",
	    "sudo apt-get -y install nginx",
	    "sudo service nginx start"
	  ]
	}
}
