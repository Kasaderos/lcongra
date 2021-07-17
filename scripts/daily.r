library('httr')
library('jsonlite')
library('lubridate')

args = commandArgs(trailingOnly=TRUE)
endpoint <- "https://api.binance.com"
url <- "/api/v3/klines"
interval <- '1d' 
#symbol <- args[2]
symbol <- 'ETHUSDT'
limit <- 1000

URL <- paste0(endpoint, url, 
		'?symbol=', symbol, 
		'&interval=', interval,
		'&limit=', limit)
r <- GET(URL)
body <- content(r, as="parsed", type="application/json") # decode json
v <- unlist(body)
df <- as.data.frame(t(matrix(as.double(v), nrow = 12)))
colnames(df) <- c('OpenTime', 'Open', 'High', 'Low', 'Close', 
		'Volume', 'CloseTime', 'QuoteAssetVolume', 
		'NumberOfTrades', 'TakerBuyBaseAV', 'TakerBuyQuoteAV',
		'Ignore')

df$OpenTime <- as_datetime(df$OpenTime / 1000)

N <- dim(df)[1]
kandle1 <- df[N-2,]
kandle2 <- df[N-1,]
kandle3 <- df[N,]

case <- 0

if (kandle1$Open > kandle1$Close && kandle2$Open < kandle2$Close && kandle3$Open < kandle3$Close) {
	case <- 1
} else if (kandle1$Open > kandle1$Close && kandle2$Open > kandle2$Close && kandle3$Open < kandle3$Close) {
	case <- 2
}

# case 1: red green green 
if (case == 1) {
	if (abs(kandle2$Open - kandle2$Close) / kandle2$Close > 0.02) {
		cat("2\n")
	} else {
	    cat("0\n")
	}
}

# case 2: red red green
if (case == 2) {
    delta <- 7
	locMin <- min(df$Close[(N-90):(N-delta)])
	if (abs(locMin - kandle3$Close) / kandle3$Close < 0.03) {
		cat("1\n")
	} else {
	    cat("-1\n")
	}
}

# empty case
if (case == 0) {
	cat("0\n")
}

