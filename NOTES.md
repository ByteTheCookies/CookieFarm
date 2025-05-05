# CookieFarm

## Roadmap

![Roadmap](images/roadmap.png)

## Todo

- [x] Setup project
  - [x] Setup client go project (makefile,logging) - @akiidjk
  - [x] Setup server go project (fiber,air,makefile,logging) - @akiidjk
  - [x] Setup frontend nuxt project (with shadcnui, tailwindcss, typescript, eslint, alias path) - @suga
  - [x] Docker base config - @akiidjk
- [x] Setup repository
  - [x] Setup .gitignore - @akiidjk @suga
  - [x] Setup security settings of repository (branch protection, code scanning, pull request review, code review, pull request approval) - @suga
- [x] Configurazione da file/shitcurl (json parsato)
- [x] Numero configurabile di thread nelle coroutine
- [x] Upgrade FE

---

## Schema

![Schema](images/schema.png)


## Feature

- [x] Dispaching flag client to server
- [x] Dispaching flag server to flag_checker
- [x] Centralized config
- [x] Authentication
- [x] Simplified config setup from web interface
- [x] Flags storing for stats and analytics
- [x] Dockerization
- [x] Runtime protocol loading using SO (Base for custom plugins)
- [x] Exploit manager

# Versions

- Bun: 1.2.9
- Go: 1.24.2
- Docker: 28.0.4


# Go code quality

- Grade .......... A+ 100.0%
- Files ................. 28
- Issues ................. 0
- gofmt ............... 100%
- go_vet .............. 100%
- gocyclo ............. 100%
- ineffassign ......... 100%
- license ............. 100%
- misspell ............ 100%

# IDEE

- [x] Aggiungere Docs (Codice e user)
- [x] Rifare log e cli di server e client
- [ ] Sistemare repo
- [x] Comando create lato client per creare template
- [ ] Altro comando cli per vedere in RealTime consumo di risorse di tutti i client e boh altre info
- [ ] Aggiungere un comando cli per vedere le flag che sono state inviate al server
- [ ] Aggiungere un bottone sulla table per aggiornarla senza ricaricare la pagina
- ? POSSIBILE REDIS PAZZO SGRAVATO (Visto che stiamo usando Docker per il server)

- [ ] Aggiungere numero client attacanti realtime displayato sulla dashboard,
- [ ] Aggiungere un display che misura ram e cpu del processo (cli - client, web - server)
- [x] Ottimizzazione exploiter per gestire servizi di macchine down etc...
- [ ] Config Reloader (hot realod of config file || button to reload)
- [ ] api/button to delete/remove the flag queue
- [ ] api/button to send flags, instead of waiting the timer
- [ ] TTL per le flag senza condizione statistiche
- [ ] client option to submit directly to gameserver
- [ ] Auto Flag



### Logging client

- Team attaccato
- Ip attaccato
- Flag prese


## Final test

- Test con infra reale
- Team separato
- Exploit dai writeup
