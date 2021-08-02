library('httr')
library('jsonlite')
library('lubridate')

args = commandArgs(trailingOnly=TRUE)
endpoint <- "https://api.binance.com"
url <- "/api/v3/klines"
interval <- args[1]
symbol <- args[2]
#interval <- "15m"
#symbol <- "BTCUSDT"
limit <- 1000

getHistory <- function(startTime, endTime){
    URL <- ''
    if (is.na(startTime) || is.na(endTime)) {
        URL <- paste0(endpoint, url, 
                      '?symbol=', symbol, 
                      '&interval=', interval,
                      '&limit=', limit)
    } else {
        URL <- paste0(endpoint, url, 
                  '?symbol=', symbol, 
                  '&interval=', interval,
                  '&startTime=', as.character(round(startTime)),
                  '&endTime=', as.character(round(endTime)),
                  '&limit=', limit)
    }
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
n <- 8
startTime <- as.numeric(Sys.time() - limit * interv * 60 * n) * 1000
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

###########################################################
# check ma25
###########################################################

interval <- '1d'
daily <- getHistory(NA, NA)


library("Mcomp")
library("smooth")
ma25 <- sma(daily$Close, h = 25)
dir <- 0
lastMonth <-
    diff(ma25$fitted[(length(ma25$fitted) - 30):length(ma25$fitted)])
s <- sum(lastMonth)

if (s > 0 && s < 100) {
    dir <- 0
} else if (s < 0) {
    dir <- -1
} else {
    dir <- 1
}

if (dir < 0) {
    cat("-1")
    quit(status=0)
} 

orig <- df
orig <- orig[1:(dim(df)[1]),2]
N <- length(orig)
ts <- orig
p <- 60
X <- matrix(ts[1:p])
for (i in 2:(N-p+1)) {
    X <- cbind(X, ts[i:(i+p-1)])
}
# norm(x, "F") x - matrix
norms <- c()
for (i in 1:(dim(X)[2]-1)) {
    if (i != N-p+1) {
        norms <- c(norms, norm(as.matrix(X[,i]-X[,N-p+1]), "I"))
    }
}
nearest <-min(norms) 
eps <- 0
step <- nearest*0.25
num_neighbors <- 0
while (num_neighbors < 3 * (p+1)){
    ws <- which(norms < (nearest + eps))
    eps <- eps + step
    num_neighbors <- length(ws)
}

if (length(norms) == 0) {
    print("neighbors == 0")
    quit(status=1)
}

Y <- X[dim(X)[1],(ws+1)]
A <- t(X[,ws])
m <- lm(Y ~ A)
#summary(m)

predict <- function(x_t, coef) {
    y <- coef[1] + t(x_t) %*% coef[2:length(coef)]
    return(y)
}
x_t <- ts[(N-p+1):(N)]
a <- as.vector(m$coefficients)
y <- predict(x_t, a)

for (i in 1:p){
    y <- predict(ts[(N-p+1):(N)], a)
    ts <- c(ts, y)
    N <- N+1
}
matplot(data.frame(ts[(length(ts)-500):length(ts)], c(df$Close[(dim(df)[1]-500+p):dim(df)[1]], rep(NA, p))), type = "l", col = c('green', 'red'),
        ylab="price", xlab="time")
abline(v=500-p, col='red')

max_price <- max(ts[(length(ts)-p):length(ts)])
min_price <- min(ts[(length(ts)-p):length(ts)])

price <- ts[length(ts)-p]
cat(paste("current", price, "max", max_price, "min", min_price, "\n"))
s <- sum(diff(ts[(length(ts)-p):length(ts)]))
cat(paste("estimate", s, "\n"))
eps <- price * 0.003
if (price + eps < max_price) { # && price - eps < min_price) {
    cat("1\n")
} else if (s < 0) {
    cat("-1\n")
} else {
    cat("0\n")
}

