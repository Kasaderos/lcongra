library('httr')
library('jsonlite')
library('lubridate')

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
getData <- function(startTime){
    initialized <- FALSE
    df <- NA
    now <- as.numeric(Sys.time()) * 1000
    while (1){
        endTime <- startTime + limit * 60 * 1000
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

startTime <- as.numeric(Sys.time())
if (interval == '1m') {
    startTime <- as.numeric(Sys.time() - limit * 60 * 8) * 1000
} else if (interval == '3m') {
    startTime <- as.numeric(Sys.time() - limit * 3 * 60 * 8 ) * 1000
}
df <- NA
if (file.exists(symbol, '.csv')) {
   startTime <- as.numeric(Sys.time() - limit * 60) * 1000
   tmpdf <- getData(startTime) 
   df <- read.csv(paste0(symbol,'.csv'), header=T)
   N <- dim(df)[1]
   lastTime <- as.numeric(df[N]$OpenTime)
   ind <- which(lastTime == as.numeric(tmpdf$OpenTime))
   if (length(ind) > 0) {
      df <- rbind(df, tmpdf[(ind+1):(dim(tmpdf)[1]),])[1:8000]
   }
} else {
   df <- getData(startTime)
}

write.csv(df, file=paste0(symbol, '.csv'))

#library('smooth')
#library('Mcomp')
#v <- sma(data_ex$Close, 25)
#orig <- v$fitted

# in df
# out ans

pdf(paste0(symbol, ".pdf"))

f <- log(df$Close[1])
orig <- diff( log(df$Close))
N <- length(orig)
ts <- orig
p <- 40
X <- matrix(rev(ts[1:p]))
for (i in 2:(N-p+1)) {
    X <- cbind(X, rev(ts[i:(i+p-1)]))
}

# norm(x, "F") x - matrix
norms <- c()
for (i in 1:(N-p)) {
    norms <- c(norms, norm(as.matrix(X[,i]-X[,N-p+1]), "F"))
}

ws <- order(norms)[1:(3*(p+1))]

#if (length(nearest) == 0) {
#    print("neighbors == 0")
#}


Y <- X[1,(ws+1)]
A <- t(X[,ws])
df2 <- as.data.frame(A)
df2 <- cbind(Y, df2)
library(randomForest)
set.seed(0)
rfManyReg <- randomForest(Y ~ ., data=df2)
#print(rfManyReg)

x_t <- ts[(N-p+1):(N)]
pr <- data.frame(t(rev(x_t)))
colnames(pr) <- colnames(df2)[2:dim(df2)[2]]
y <- predict(rfManyReg, pr)

#print(data_ex[N+1,2])
#print(y)
#N <- p
for (i in 1:p){
    x_t <- ts[(N-p+1):(N)]
    pr <- data.frame(t(rev(x_t)))
    colnames(pr) <- colnames(df2)[2:dim(df2)[2]]
    y <- predict(rfManyReg, pr)
    ts <- c(ts, y)
    N <- N+1
}
retback <- function(ts, f){
   t <- c(f)
   for (i in 1:length(ts)){
       t <- c(t, ts[i]+t[i])
   } 
   t
}
ts <- exp(retback(ts, f))
matplot(data.frame(ts[(length(ts)-500):length(ts)], c(df$Close[(dim(df)[1]-500+p):dim(df)[1]], rep(NA, p))), type = "l", col = c('green', 'red'),
        ylab="price", xlab="time")
abline(v=500-p, col='red')

max_price <- max(ts[(length(ts)-p):length(ts)])
min_price <- min(ts[(length(ts)-p):length(ts)])

price <- ts[length(ts)-p+1]
eps <- price * 0.003
if (price < min_price + eps && price < max_price - eps){ # && price - eps < min_price) {
    cat(paste("1", price, min_price, max_price))
} else if (price > min_price + eps) {
    cat(paste("-1", price, min_price, max_price))
} else {
    cat(paste("0", price, min_price, max_price))
}
