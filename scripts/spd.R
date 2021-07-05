library('httr')
library('jsonlite')
library('lubridate')
library("Mcomp")
library("smooth")

args = commandArgs(trailingOnly=TRUE)
endpoint <- "https://api.binance.com"
url <- "/api/v3/klines"
interval <- args[1]
symbol <- args[2]

limit <- 1000

getHistory <- function(startTime, endTime){
    URL <- paste0(endpoint, url, 
                  '?symbol=', symbol, 
                  '&interval=', interval,
                  '&startTime=', as.character(round(startTime)),
                  '&endTime=', as.character(round(endTime)),
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
    return(df)
}

# in ms
getData <- function(startTime, interv){
    initialized <- FALSE
    df <- NA
    now <- as.numeric(Sys.time()) * 1000
    while (1){
        endTime <- startTime + limit * interv * 60 * 1000
        if (initialized) {
            df <- getHistory(startTime, endTime)
        } else {
            df2 <- getHistory(startTime, endTime)
            df <- rbind(df, df2)
        }
        if (startTime >= now) {
            df <- df[-1,]
            return(df)
        }
        startTime <- endTime
    }
}

if (interval != "15m") {
    quit(status = 1)
}

interv <- 15
startTime <- as.numeric(Sys.time() - limit * interv * 60) * 1000
df <- NA
if (file.exists(symbol, '.csv')[1]) {
   startTime <- as.numeric(Sys.time() - limit * interv * 60) * 1000
   tmpdf <- getData(startTime) 
   df <- read.csv(paste0(symbol,'.csv'), header=T)
   N <- dim(df)[1]
   lastTime <- as.numeric(df[N]$OpenTime)
   ind <- which(lastTime == as.numeric(tmpdf$OpenTime))
   if (length(ind) > 0) {
      df <- rbind(df, tmpdf[(ind+1):(dim(tmpdf)[1]),])[1:8000]
   }
} else {
   df <- getData(startTime, interv)
}

write.csv(df, file=paste0(symbol, '.csv'))

pdf(paste0(symbol, ".pdf"))

m <- sma(df$Close, 25)
matplot(data.frame(m$fitted, df$Close), type = "l", col = c('red', 'green'),
        ylab="price", xlab="time")
p <- 5
N <- length(m$fitted) - 1
dr <- diff(m$fitted)[(N-p):N]
lastPrice <- df$Close[length(df$Close)]

if (length(which(dr > 0)) == length(dr) && mean(dr) >= lastPrice * 0.001) {
    cat("1")
} else if (length(which(dr < 0)) == length(dr)) {
    cat("-1")
} else {
    cat("0")
}