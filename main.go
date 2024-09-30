package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	guildID  string = "813458127310946364"
	commands        = []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Ответит Pong!",
		},
		{
			Name:        "echo",
			Description: "Повторит ваш ввод",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "текст для повторения",
					Required:    true,
				},
			},
		},
	}
)

// main - Точка входа в программу.
//
// main - главная функция программы, которая:
//
// 1. Читает токен бота из переменной окружения BOT_TOKEN.
// 2. Создает сессию бота.
// 3. Добавляет обработчик событий interactionHandler.
// 4. Открывает соединение с сервером Discord.
// 5. Создает все команды, описанные в переменной commands.
// 6. Ждет, пока не будет нажата клавиша CTRL+C.
// 7. Удаляет все созданные команды.
// 8. Закрывает соединение с сервером Discord.
func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		fmt.Println("Токен бота не найден в перененной окружения BOT_TOKEN")
		return
	}

	var err error
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Ошбка при создании сессии:", err)
		return
	}

	dg.AddHandler(interactionHandler)

	err = dg.Open()
	if err != nil {
		fmt.Println("Ошибка при открытии соединения:", err)
		return
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	for i, cmd := range commands {
		rc, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, cmd)
		if err != nil {
			fmt.Println("Ошибка при создании команды:", cmd.Name, err)
		} else {
			registeredCommands[i] = rc
		}
	}

	fmt.Println("Бот запущен. Нажмите CTRL+C для выхода.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	for _, cmd := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, guildID, cmd.ID)
		if err != nil {
			fmt.Println("Ошибка при удалении команды:", cmd.Name, err)
		}
	}

	dg.Close()
}

// interactionHandler - Обработчик события InteractioCreate, возникает при любом
// взаимодействии с ботом, например, при вводе команды.
//
// s - сессия бота.
// i - данные события InteractioCreate.
//
// interactionHandler обрабатывает команды "ping" и "echo". Команда "ping" отправляет
// ответ "Pong!", а команда "echo" повторяет введенный текст.
func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		data := i.ApplicationCommandData()
		switch data.Name {
		case "ping":
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
		case "echo":
			content := data.Options[0].StringValue()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		}
	}
}
