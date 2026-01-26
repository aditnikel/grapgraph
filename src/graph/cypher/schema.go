package cypher

const CreateUserIndex = `CREATE INDEX ON :User(user_id)`
const CreateMerchantIndex = `CREATE INDEX ON :Merchant(merchant_id_mpan)`
const CreateDeviceIndex = `CREATE INDEX ON :Device(device_id)`
const CreatePaymentMethodIndex = `CREATE INDEX ON :PaymentMethod(payment_method)`
const CreateBankIndex = `CREATE INDEX ON :Bank(issuing_bank)`
const CreateWalletIndex = `CREATE INDEX ON :Wallet(wallet_address)`
const CreateExchangeIndex = `CREATE INDEX ON :Exchange(exchange)`
