FROM golang

WORKDIR /usr/src/discorgi

COPY . .

RUN go get github.com/bwmarrin/discordgo

CMD ["go", "run", "main.go"]