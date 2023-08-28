terraform {
  require_version = ">= 1.5.1"
}

locals {
  foo = "lorem ipsum dolor sit amet"
}

variable "bar" {
  type    = int
  default = 1024
}

variable "foo" {
  type        = string
  default     = "FOO"
  description = <<EOD
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec vel egestas dolor, nec dignissim metus. Donec augue elit, rhoncus ac sodales id, porttitor vitae est. Donec laoreet rutrum libero sed pharetra.

 Donec vel egestas dolor, nec dignissim metus. Donec augue elit, rhoncus ac sodales id, porttitor vitae est. Donec laoreet rutrum libero sed pharetra. Duis a arcu convallis, gravida purus eget, mollis diam.
EOD
}


resource "null_resource" "foo" {}

resource "null_resource" "bar" {
  count = 2

  triggers = {
    foo = null_resource.foo.id
  }

  provisioner "local-exec" {
    inline = [
      "hostname",
      "uname -a",
    ]
  }
}

resource "null_resource" "quux" {
  count = 3

  triggers = {
    foo = null_resource.foo.id
  }

  provisioner "remote-exec" {
    inline = [
      "hostname",
      "uname -a",
    ]
  }
}

data "bogus" "foo" {
  input "a" {
    value = 1
  }
  input "b" {
    value = 2
  }

  outputs {
    safe "a" { value = "a" }
    unsafe "b" { value = "b" }
  }
}

