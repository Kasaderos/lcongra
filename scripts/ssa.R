library('httr')
library('jsonlite')
library('lubridate')

args = commandArgs(trailingOnly=TRUE)
endpoint <- "https://api.binance.com"
url <- "/api/v3/klines"
interval <- args[1]
symbol <- args[2]
position <- args[3]

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
if (file.exists(symbol, '.csv')) {
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

library('Rssa')
len <- 1 
N <- dim(df)[1]
f <- df$Close
s <- ssa(f)
rfor <- rforecast(s, groups = list(c(1,4), 1:4), len = len, only.new=FALSE)
matplot(data.frame(c(f, rep(NA, len)), rfor$F2), type = "l", col = c('red', 'green', 'blue'),
        ylab="price", xlab="time")
abline(v=N, col='red')

p <- 25
N <- length(rfor$F2)
lastPrice <- rfor$F2[(N-len)]
lastPoints <- lastPrice - rfor$F2[(N-p-len):(N-1-len)]
cat(lastPoints)
cat("\n")

if (position == "Close") {
    eps <- 3
    ind <- c()
    for (i in (eps+2):(length(lastPoints)-eps)) {
        if (lastPoints[i-1] > 0 && lastPoints[i] < 0) {
            ind <- c(ind, i-1)
        }
    }
    if (length(ind) == 0 && length(which(lastPoints > 0)) == length(lastPoints)) {
        cat("1")
        quit(status = 0)
    }
    if (length(ind) == 0 && length(which(lastPoints < 0)) == length(lastPoints)) {
        cat("-1")
        quit(status = 0)
    }
    if (length(ind) == 0) {
        cat("1")
        quit(status = 0)
    }
    for (i in 1:length(ind)) {
        ps <- length(which(lastPoints[(ind[i]-eps+1):(ind[i])] > 0))
        ns <- length(which(lastPoints[(ind[i]+1):(ind[i]+eps)] < 0))
        if (ps == eps && ns == eps) {
            cat("0")
            break
        }
    }
} else {
    eps <- 3
    ind <- c()
    for (i in (eps+2):(length(lastPoints)-eps)) {
        if (lastPoints[i-1] < 0 && lastPoints[i] > 0) {
            ind <- c(ind, i-1)
        }
    }
    if (length(ind) == 0 && length(which(lastPoints < 0)) == length(lastPoints)) {
        cat("-1")
        quit(status = 0)
    }
    if (length(ind) == 0 && length(which(lastPoints > 0)) == length(lastPoints)) {
        if (lastPoints[1] > lastPrice * 0.0027) {
            cat("1")
        } else {
            cat("0")
        }
        quit(status = 0)
    }
    if (length(ind) == 0) {
        cat("0")
        quit(status = 0)
    }

    for (i in 1:length(ind)) {
        ps <- length(which(lastPoints[(ind[i]-eps+1):(ind[i])] < 0))
        ns <- length(which(lastPoints[(ind[i]+1):(ind[i]+eps)] > 0))
        if (ps == eps && ns == eps) {
            cat("0")
            break
        }
    }
}
