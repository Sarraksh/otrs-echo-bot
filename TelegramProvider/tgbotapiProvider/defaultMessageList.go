package tgbotapiProvider

const (
	invalidCommandResponse string = `Неверная команда. Для отображения списка команд используйте /help .`
	helpCommandResponse    string = `Список доступных команд:

Управление подписками на события
/subscribeTeam1
/subscribeTeam2
/subscribeTeam3
/unsubscribeTeam1
/unsubscribeTeam2
/unsubscribeTeam3

Указание своих имени и фамилии
/firstName Имя
/lastName Фамилия

Вывод данного сообщения
/help
`
	startCommandResponse string = `Для начала работы с ботом пожалуйста оформите подписку на события одну из команд:
/subscribeTeam1
/subscribeTeam2
/subscribeTeam3

Оформленную подписку можно отменить в любой момент аналогичными командами:
/unsubscribeTeam1
/unsubscribeTeam2
/unsubscribeTeam3

Для автоматического получения сообщений по всем событиям для дежурных пожалуйста укажите свои имя и фамилию с помощью следующих команд:
/firstName Имя
/lastName Фамилия`
	invalidFirstNameResponse string = `Имя должно содержать только русские буквы.`
	invalidLastNameResponse  string = `Фамилия должна содержать только русские буквы.`
)
