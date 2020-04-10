# apizza configuration

## Config Fields
#### name
The name field will be the name sent to Dominos whenever an order is sent.

#### email
This is the email that will be sent to Dominos whenever an order is sent. Email is one of the identifiers that Dominos uses to keep track of people so if you set the email field (and the phone field), Dominos will give you one credit towards a free pizza.

#### phone
The phone field will also be used when sending an order to Dominos. As mentioned in the [email](#email) section, Dominos uses phone numbers (and email) to identify people and give them credit toward free pizza.

#### address
The address config field is currently being phased out in. Use `apizza address` to add an address instead. The `street` subfield should include your street number and street name. The rest of the address subfields should be self-explanatory.

#### default-address-name
This field sets the default value used for the `--address, -A` flag. The value of this field should be the name of one of the addresses stored when `apizza address --new` is executed and completed.

#### card
The card field will include the card number and expiration date for a payment when ordering. The date should be in the format `mm/yy`.

#### service
This field should be either "Carryout" or "Delivery". "Delivery" if you want you food to be delivered and "Carryout" if you want to go pick you food up in person.
