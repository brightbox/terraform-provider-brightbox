provider "brightbox" {
}

resource "brightbox_server" "first_server" {
	image = "img-m8vud"
	name = "Test terraform server"
}
