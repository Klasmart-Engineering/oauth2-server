resource "aws_dynamodb_table" "authentication" {
  name         = "authentication"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"
  range_key    = "sk"

  global_secondary_index {
    name = "gsi-1"

    projection_type = "ALL"

    hash_key = "sk"
  }

  attribute {
    name = "pk"
    type = "S"
  }

  attribute {
    name = "sk"
    type = "S"
  }
}
