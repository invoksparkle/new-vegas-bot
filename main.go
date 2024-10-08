package main

import (
	"fmt"
	"os"
	"os/exec"
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
		{
			Name:        "radio",
			Description: "Проигрывает радио из Fallout New Vegas",
		},
	}
)

// main - функция, которая является точкой входа программы.
// Она инициализирует бота, регистрирует команды, запускает
// обработчик событий и останавливает бота при SIGINT, SIGTERM
// или CTRL+C.
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
		case "radio":
			voiceChannelID := "834079206186024981"
			err := playRadio(s, i.GuildID, voiceChannelID)
			if err != nil {
				fmt.Println("Ошибка при проигрывании радио:", err)
			}
		}
	}
}

// playRadio - проигрывает радио из Fallout New Vegas в указанный голосовой
// канал.
//
// s - сессия бота.
// guildID - ID сервера.
// voiceChannelID - ID голосового канала, в котором будет проигрываться радио.
//
// Возвращает ошибку, если возникла проблема при проигрывании радио.
func playRadio(s *discordgo.Session, guildID, voiceChannelID string) error {
	vc, err := s.ChannelVoiceJoin(guildID, voiceChannelID, false, true)
	if err != nil {
		return fmt.Errorf("не удалось присоединиться к голосовому каналу: %v", err)
	}

	cmd := exec.Command("ffmpeg", "-i", "https://fallout.fm:8444/falloutfm3.ogg", "-f", "s16le", "-ar", "48000", "-ac", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error retrieving audio stream: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting ffmpeg: %v", err)
	}

	vc.Speaking(true)
	defer vc.Speaking(false)

	buff := make([]byte, 960*2)

	for {
		_, err := stdout.Read(buff)
		if err != nil {
			break
		}
		vc.OpusSend <- buff
	}

	return nil
}
