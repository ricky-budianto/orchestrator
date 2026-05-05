package utils

const (
	// HTTP status code
	HTTPOk                  = 200
	HTTPCreated             = 201
	HTTPAccepted            = 202
	HTTPBadRequest          = 400
	HTTPUnauthorized        = 401
	HTTPForbidden           = 403
	HTTPNotFound            = 404
	HTTPInternalServerError = 500

	// Application Constant
	AppName    = "CES Customer"
	AppVersion = "0.0.1"

	// Message
	MessageSuccess       = "success"
	MessageNoDataUpdated = "No data updated"
	MessageNoCacheData   = "No cache data"

	// MODULE
	ModuleCustomer         = "customer"
	ModuleUser             = "user"
	ModuleRegistration     = "registration"
	ModuleSyncTransferList = "sync_transfer_list_attempt"
	ModuleOTP              = "MULTI_DEVICE_OTP"

	// API Path
	PathService               = "/api/ibridge"
	PathV1                    = "/v1"
	PathV2                    = "/v2"
	PathBase                  = ""
	PathID                    = "/:id"
	ParamID                   = "/:param_id"
	PathWorkflowAudit         = "/workflow-audit"
	PathWorkflowConfiguration = "/workflow-configuration"
	PathWorkflowState         = "/workflow-state"
	PathWorkflowHistory       = "/workflow-history"
	PathWorkflow              = "/workflow"
	PathOrchestrate           = "/orchestrate"
	PathElasticLog            = "/logs"

	PathLogin            = "/login"
	PathLogout           = "/logout"
	PathStatus           = "/status"
	PathIDPassword       = "/:id/password"
	PathRequest          = "/request"
	PathChange           = "/change"
	PathRead             = "/read"
	PathSave             = "/save"
	PathVerify           = "/verify"
	PathGenerate         = "/generate"
	PathHash             = "/hash"
	PathSubmit           = "/submit"
	PathMaster           = "/master"
	PathRegion           = "/region"
	PathRegistration     = "/registration"
	PathCardless         = "/cardless"
	PathCancel           = "/cancel"
	PathCashout          = "/cashout"
	PathValidate         = "/validate"
	PathSignature        = "/signature"
	PathCrypto           = "/crypto"
	PathEbranch          = "/ebranch"
	PathSyncTransferList = "/sync-transfer-list"
	PathRestrictions     = "/restrictions"

	// account params
	PathAccountParams = "/account-params"

	// flags
	PathFlags = "/flags"

	// customer
	PathThoughtMachine = "/thought-machine"

	// SYSTEM PARAMETER
	SPCustomerTemporaryPassword         = "customer_temporary_password"
	SPGroupSMTP                         = "smtp"
	SPSMTPSubjectRequestChangePassword  = "smtp_subject_request_change_password"
	SPSMTPTemplateRequestChangePassword = "smtp_template_request_change_password"

	// myBCAS parameter
	PathMyBCASParameter                 = "/mybcas/parameter"
	PathMyBCASParameterMobileVersion    = "/mobile-app/:os/version"
	PathMyBCASParameterSknRtgsCities    = "/skn-rtgs/cities"
	PathMyBCASParameterFlazzCredentials = "/flazz/credentials"
	PathMyBCASParameterID               = "/:id"

	// Product Params Update
	PathProductParams    = "/product-params"
	PathGetProductParams = "/product-params/:id"

	// sub Product Parameter
	PathSubProductParameter    = "/sub-product-parameter"
	PathGetSubProductParameter = "/sub-product-parameter/:id"

	// Wadiah Fees
	PathWadiahFees        = "/wadiah/fees"
	PathWadiahFeesReserve = "/wadiah/fees-reserve"
	PathWadiahFeePayment  = "/wadiah/fees/payment"

	// Gold Financing Account Details
	PathGoldFinancingAccountDetails = "/gold-financing/account/details"

	// Crypto
	PathCryptoGetPublicKey = "/get-public-key"

	// Cardless Response Message
	CWFailedCancel = "Mohon maaf, transaksi Cardless anda telah berhasil di Tarik Tunai dan tidak dapat dibatalkan"

	// Mapping Document
	MDConcernInput = "document_checklist"
	MDConcernData  = "concerns"
	MDCIF          = "cif"
	MDName         = "username"
	MDTimeLocation = "Asia/Jakarta"
	MDTimeFormat   = "Monday, 02 January 2006"
	MDChecked      = "checked"

	// Name Day in Indonesian
	IndDay01 = "Senin"
	IndDay02 = "Selasa"
	IndDay03 = "Rabu"
	IndDay04 = "Kamis"
	IndDay05 = "Jumat"
	IndDay06 = "Sabtu"
	IndDay07 = "Minggu"

	// Name Day in English
	EngDay01 = "Monday"
	EngDay02 = "Tuesday"
	EngDay03 = "Wednesday"
	EngDay04 = "Thursday"
	EngDay05 = "Friday"
	EngDay06 = "Saturday"
	EngDay07 = "Sunday"

	// Name Month in Indonesian
	IndMonth01 = "Januari"
	IndMonth02 = "Februari"
	IndMonth03 = "Maret"
	IndMonth04 = "April"
	IndMonth05 = "Mei"
	IndMonth06 = "Juni"
	IndMonth07 = "Juli"
	IndMonth08 = "Agustus"
	IndMonth09 = "September"
	IndMonth10 = "Oktober"
	IndMonth11 = "November"
	IndMonth12 = "Desember"

	IndMo01 = "Jan"
	IndMo02 = "Feb"
	IndMo03 = "Mar"
	IndMo04 = "Apr"
	IndMo05 = "Mei"
	IndMo06 = "Jun"
	IndMo07 = "Jul"
	IndMo08 = "Agt"
	IndMo09 = "Sep"
	IndMo10 = "Okt"
	IndMo11 = "Nov"
	IndMo12 = "Des"

	// Name Month in English
	EngMonth01 = "January"
	EngMonth02 = "February"
	EngMonth03 = "March"
	EngMonth04 = "April"
	EngMonth05 = "May"
	EngMonth06 = "June"
	EngMonth07 = "July"
	EngMonth08 = "August"
	EngMonth09 = "September"
	EngMonth10 = "October"
	EngMonth11 = "November"
	EngMonth12 = "December"

	EngMo01 = "Jan"
	EngMo02 = "Feb"
	EngMo03 = "Mar"
	EngMo04 = "Apr"
	EngMo05 = "May"
	EngMo06 = "Jun"
	EngMo07 = "Jul"
	EngMo08 = "Aug"
	EngMo09 = "Sep"
	EngMo10 = "Oct"
	EngMo11 = "Nov"
	EngMo12 = "Dec"

	// Mapping Transfer List
	MTLTypeTransfer                          = "transfer"
	MTLTypePayment                           = "payment"
	MTLTypePurchase                          = "purchase"
	MTLRegex                                 = `^[0-9]+$`
	MTLPaymentTrxType                        = "payment_transaction_type"
	MTLBRISCode                              = "422"
	MTLBSICode                               = "451"
	MTLTRFH2HCode                            = "014"
	MTLTRFH2H                                = "transfer_bca_outgoing"
	MTLBCASBankCode                          = "A01"
	MTLXIPSwiftCode                          = "SYCAIDJ1"
	MTLDateLayout                            = "2006-01-02"
	MTLBCASParameterID                       = "biller_product_group"
	MTLBillerPayment                         = "biller_payment"
	MTLBillerPaymentCode                     = "82"
	MTLBillerPurchase                        = "biller_purchase"
	MTLBillerPurchaseCode                    = "92"
	MTLBillerPaymentType                     = "mobile_postpaid"
	MTLBillerPurchaseCreditType              = "mobile_credit"
	MTLBillerPurchaseInternetQuotaType       = "mobile_internet_quota"
	MTLBillerPurchaseInternetQuotaTypeCode01 = "PDX"
	MTLBillerPurchaseInternetQuotaTypeCode02 = "PDA"
	MTLBillerProductGroupAxis                = "AXS"
	MTLBillerPaymentIndihomeCode             = "89"
	MTLBillerProductGroupIndihome            = "PTLKM"
	MTLBillerProductGroupIndosat             = "ISAT"
	MTLBillerProductGroupSmartfren           = "SMRT"
	MTLBillerProductGroupTelkomsel           = "TSEL"
	MTLBillerProductGroupTri                 = "THRE"
	MTLBillerProductGroupXL                  = "XL"
	MTLBillerProductGroupXLPrio              = "XLPrio"
	MTLBillerPLNPostpaidCode                 = "81"
	MTLBillerPLNPrepaidCode                  = "9a"
	MTLBillerPLNNonTaglisCode                = "8b"
	MTLBillerPDAMCode                        = "80"
	MTLBillerPDAMProductID01                 = "AER"
	MTLBillerPDAMProductID02                 = "PAL"
	MTLBillerPDAMProductID03                 = "PAMBDG"
	MTLBillerPDAMProductID04                 = "PAMSBY"
	MTLBillerPDAMProductID05                 = "PAMSMG"

	// OTP Status
	OTPActive  = "ACTIVE"
	OTPExpired = "EXPIRED"
	OTPUsed    = "USED"

	// Response Code
	RCSuccess                      = "0000"
	RCNoDataUpdated                = "0001"
	RCRequestHasBeenSent           = "0002"
	RCUnknownError                 = "1001"
	RCSystemError                  = "1002"
	RCDatabaseError                = "1003"
	RCFileSystemError              = "1004"
	RCThirdPartySystemError        = "1005"
	RCConnectionTimeout            = "1006"
	RCDataNotFound                 = "1007"
	RCDuplicateData                = "1008"
	RCImmutableData                = "1009"
	RCNotAuthorizedAccess          = "1010"
	RCIInvalidCredential           = "1011"
	RCUserIsLoggedIn               = "1012"
	RCInvalidLoginSession          = "1013"
	RCUnsupportBodyType            = "1014"
	RCMissingParameter             = "1015"
	RCInvalidInputFormat           = "1016"
	RCUploadFileFailed             = "1017"
	RCOTPHasBeenSent               = "1018"
	RCOTPHasExpired                = "1019"
	RCOTPInvalid                   = "1020"
	RCPINHasExpired                = "1021"
	RCPINInvalid                   = "1022"
	RCAccountNotFullySetup         = "1023"
	RCAccountDisabled              = "1024"
	RCBadRequest                   = "1025"
	RCInvalidSignature             = "1026"
	RCOTPAttempBlocked             = "1027"
	RCOTPMaxAttemp                 = "1028"
	RCInvalidCode                  = "1029"
	RCCodeHasExpired               = "1030"
	RCInvalidInputData             = "1031"
	RCValidationProcessNotComplete = "1032"
	RCUnregistredDeviceID          = "1033"
	RCComunicateWithIbridge        = "1034"
	RCOTPAttempBlockedPermanent    = "1035"
	RCOTPAttempBlockedByNumber     = "1036"
	RCInternalServerError          = "1037"
	RCAuthMaxAttemp                = "1038"
	RCAuthAttempBlockedPermanent   = "1039"
	RCOTPMaxSend                   = "1059"
	RCOTPMaxResend                 = "1060"
	RCOTPMaxVerify                 = "1061"
	RCRedisConnection              = "1062"
	RCMaritalStatusInvalid         = "1063"

	// Other
	CoreBankingLegacy = "legacy"
	CoreBankingTM     = "TM"

	// MyBCAS Parameter
	MyBCASParameterGenerateAccountPrefixID         = "generate_account_prefix"
	MyBCASParameterAccountPrefixesID               = "account_prefixes"
	MyBCASParameterTMProductIDs                    = "tm_product_ids"
	MyBCASParameterMobileAppVersionID              = "mobile_app_version"
	MyBCASParameterGoldFinancingCustomerTypes      = "gold_financing_customer_types"
	MyBCASParameterAllowedProductIDsAsSourceOfFund = "allowed_product_ids_as_source_of_fund"
	MyBCASParameterSknRtgsCities                   = "skn_rtgs_cities"
	MyBCASParameterFlazzCredentials                = "flazz_credentials"

	// Account status
	AccountStatusOpen         = "active"
	AccountStatusClosed       = "inactive"
	AccountStatusDormant      = "dormant"
	AccountStatusPendingClose = "pending_close"
	AccountStatusUnknown      = "unknown"

	// TM Account Types Based on Product Tags
	TMProductTagSavings     = "SAVINGS"
	TMProductTagFinancing   = "FINANCING"
	TMProductTagTimeDeposit = "TIME_DEPOSIT"
	TMProductTagSavingPlan  = "SAVING_PLAN"

	// ISI Account Types
	ISIProductTypeTahapan  = "T"
	ISIProductTypeGiro     = "G"
	ISIProductTypeRencana  = "R"
	ISIProductTypeDeposito = "D"
	ISIProductTypeLoan     = "L"

	CorrelationIDKey string = "correlation_id"
)

var MonthENtoID = map[string]string{
	EngMonth01: IndMonth01,
	EngMonth02: IndMonth02,
	EngMonth03: IndMonth03,
	EngMonth04: IndMonth04,
	EngMonth05: IndMonth05,
	EngMonth06: IndMonth06,
	EngMonth07: IndMonth07,
	EngMonth08: IndMonth08,
	EngMonth09: IndMonth09,
	EngMonth10: IndMonth10,
	EngMonth11: IndMonth11,
	EngMonth12: IndMonth12,
}

var MoENtoID = map[string]string{
	EngMo01: IndMo01,
	EngMo02: IndMo02,
	EngMo03: IndMo03,
	EngMo04: IndMo04,
	EngMo05: IndMo05,
	EngMo06: IndMo06,
	EngMo07: IndMo07,
	EngMo08: IndMo08,
	EngMo09: IndMo09,
	EngMo10: IndMo10,
	EngMo11: IndMo11,
	EngMo12: IndMo12,
}
