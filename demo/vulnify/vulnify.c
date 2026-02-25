#include "vulnify.h"

struct playlist_t
{
	char name[PLAYLIST_NAME_LENGTH];
	char description[PLAYLIST_DESCRIPTION_LENGTH];
	char songs[MAX_SONGS][SONG_NAME_LENGTH];
	uint8_t saved_songs;
};

struct user_t
{
	char username[USERNAME_LENGTH];
	char password[PASSWORD_LENGTH];
	uint8_t saved_playlists;
};

user* g_logged = NULL;


int main()
{
	setvbuf(stdin, NULL, _IONBF, 0);
	setvbuf(stdout, NULL, _IONBF, 0);
	setvbuf(stderr, NULL, _IONBF, 0);
	setlocale(LC_ALL, "");

	g_logged = malloc(sizeof(user));

	main_menu();

	free(g_logged);
	g_logged = NULL;

	return 0;
}

uint8_t _user_exists(char* name, char* path)
{/*{{{*/
	name[strcspn(name, "\n")] = '\0';
	snprintf(path, USER_PATH_LENGTH, "data/%s", name);
	return access(path, F_OK) == 0;
}/*}}}*/

uint8_t _playlist_exists(char* name, char* path)
{/*{{{*/
	snprintf(path, PLAYLIST_PATH_LENGTH, "data/%s/%s", g_logged->username, name);
	return access(path, F_OK) == 0;
}/*}}}*/

void _load_user(char* path)
{/*{{{*/
	snprintf(path + strcspn(path, "\n"), 6, "/info");
	FILE* f = fopen(path, "rb");
	fread(g_logged, sizeof(user), 1, f);
	fclose(f);
}/*}}}*/

void _save_user(char* path, user* usr)
{/*{{{*/
	snprintf(path + strlen(path), USER_PATH_LENGTH + 6, "/info");
	FILE* uf = fopen(path, "w+b");
	fwrite(usr, sizeof(user), 1, uf);
	fclose(uf);
}/*}}}*/

uint8_t _check_string(char* string)
{/*{{{*/
	string[strcspn(string, "\n")] = '\0';

	while (*string != '\0')
	{
		if (!isalnum(*string) && *string != ' ') break;
		++string;
	}

	return *string == '\0';
}/*}}}*/

void generate_safe_password(char* username, char* password)
{/*{{{*/
	uint32_t val = 1, tmp;
	for (uint8_t j = 0; j < (uint8_t)(strlen(username) / 4.f) * 4; j += 4)
	{
		tmp = username[j] * (256 * 256 * 256) + username[j + 1] * (256 * 256) + username[j + 2] * 256 + username[j + 3];
		val = (0xDEAD * (val + tmp) + 0xBEEF) & ((1L << 32) - 1);
	}
	snprintf(password, PASSWORD_LENGTH, "%u", val);
}/*}}}*/

void main_menu()
{/*{{{*/
	puts("Vulnify (version: 2.0.1)\n");
	uint8_t choice, run = 0;

	do
	{
		run = 0;

		puts("\nSelect an option");
		puts("1. Register");
		puts("2. Login");
		puts("0. Exit");
		putchar('>'); putchar(' ');

		choice = getchar(); getchar();
		putchar('\n');

		switch (choice)
		{
			case '1':
				if (!register_user()) run = 1;
				break;
			case '2':
				if (!login()) run = 1;
				break;
			default:
				run = 2;
				break;
		}
	} while (run == 1);

	if (run == 2) return;

	user_menu();
}/*}}}*/

void user_menu()
{/*{{{*/
	uint8_t choice, run = 0;

	do
	{
		puts("\nSelect an option");
		puts("1. Create a new playlist");
		puts("2. Inspect your playlists");
		puts("3. Play a random song");
		puts("0. Exit");
		putchar('>'); putchar(' ');

		choice = getchar(); getchar();

		switch (choice)
		{
			case '1':
				create_playlist();
				run = 1;
				break;
			case '2':
				inspect_playlists();
				run = 1;
				break;
			case '3':
				play_random_song();
				run = 1;
				break;
			default:
				run = 0;
				break;
		}
	} while (run);
}/*}}}*/

uint8_t register_user()
{/*{{{*/
	char username[USERNAME_LENGTH];
	char password[PASSWORD_LENGTH];
	char path[USER_PATH_LENGTH];

	while (1)
	{
		puts("Insert username");
		putchar('>'); putchar(' ');

		if (fgets(username, USERNAME_LENGTH, stdin) == NULL || strlen(username) < 1 || !_check_string(username))
		{
			puts("[ERROR] Invalid username");
			continue;
		}
		if (_user_exists(username, path))
		{
			puts("[ERROR] Username already exists");
			continue;
		}
		mkdir(path, 0777);

		break;
	}
	username[strcspn(username, "\n")] = '\0';

	puts("Insert password [Empty to generate a safe password automatically]");
	putchar('>'); putchar(' ');

	if (fgets(password, PASSWORD_LENGTH, stdin) == NULL || !_check_string(password))
	{
		puts("[ERROR] Invalid password");
		return 0;
	}
	if (password[0] == '\0')
	{
		generate_safe_password(username, password);
		printf("Your password is: %s\n", password);
	}
	else password[strcspn(password, "\n")] = '\0';


	user* usr = malloc(sizeof(user));

	strcpy(usr->username, username);
	strcpy(usr->password, password);
	usr->saved_playlists = 0;

	_save_user(path, usr);

	g_logged = usr;

	puts("\nUser succesfully registered");

	return 1;
}/*}}}*/

uint8_t login()
{/*{{{*/
	char username[USERNAME_LENGTH];
	char password[PASSWORD_LENGTH];
	char path[USER_PATH_LENGTH];

	while (1)
	{
		puts("Insert username");
		putchar('>'); putchar(' ');

		if (fgets(username, USERNAME_LENGTH, stdin) == NULL || !_user_exists(username, path))
		{
			puts("[ERROR] Username not found\n");
			continue;
		}

		break;
	}
	username[strcspn(username, "\n")] = '\0';

	_load_user(path);

	puts("Insert password");
	putchar('>'); putchar(' ');

	if (fgets(password, PASSWORD_LENGTH, stdin) == NULL || strlen(password) < 1 || strncmp(g_logged->password, password, strlen(password) - 1) != 0)
	{
		puts("[ERROR] Invalid password");
		return 0;
	}

	puts("\nLogin successful");

	return 1;
}/*}}}*/

void create_playlist()
{/*{{{*/
	char name[PLAYLIST_NAME_LENGTH];
	char description[PLAYLIST_DESCRIPTION_LENGTH];
	char song[SONG_NAME_LENGTH];
	char songs[MAX_SONGS][SONG_NAME_LENGTH];
	char path[PLAYLIST_PATH_LENGTH];
	uint8_t num_songs;

	while (1)
	{
		puts("\nInsert playlist name");
		putchar('>'); putchar(' ');

		if (!fgets(name, PLAYLIST_NAME_LENGTH, stdin) || strlen(name) < 1 || !_check_string(name))
		{
			puts("[ERROR] Invalid name");
			continue;
		}
		if (_playlist_exists(name, path))
		{
			puts("[ERROR] Playlist with that name already exists");
			continue;
		}

		break;
	}
	name[strcspn(name, "\n")] = '\0';

	while (1)
	{
		puts("\nInsert description");
		putchar('>'); putchar(' ');

		if (!fgets(description, PLAYLIST_NAME_LENGTH, stdin) || !_check_string(description))
		{
			puts("[ERROR] Invalid description");
			continue;
		}

		break;
	}
	description[strcspn(description, "\n")] = '\0';

	while (1)
	{
		puts("\nHow many songs do you want to insert?");
		putchar('>'); putchar(' ');

		scanf("%hhu%*c", &num_songs);
	 	if (num_songs >= MAX_SONGS || num_songs == 0)
		{
			printf("[ERROR] Select a positive number smaller than %hhd\n", MAX_SONGS);
			continue;
		}

		break;
	}

	puts("\nInsert songs");
	uint8_t i = 0;
	while (i < num_songs)
	{
		printf("Song %hhd: ", i);

		if (!fgets(song, SONG_NAME_LENGTH, stdin) || !_check_string(song))
		{
			puts("[ERROR] Invalid song name");
			continue;
		}
		song[strcspn(song, "\n")] = '\0';

		strcpy(songs[i++], song);
	}
	putchar('\n');

	playlist* pl = malloc(sizeof(playlist));

	strcpy(pl->name, name);
	strcpy(pl->description, description);
	pl->saved_songs = num_songs;

	for (i = 0; i < num_songs; ++i)
		strcpy(pl->songs[i], songs[i]);

	g_logged->saved_playlists++;

	FILE* plf = fopen(path, "w+b");
	fwrite(pl, sizeof(playlist), 1, plf);
	fclose(plf);

	free(pl);
	pl = NULL;

	puts("Playlist successfully created");
}/*}}}*/

void inspect_playlists()
{/*{{{*/
	char path[USER_PATH_LENGTH], fpath[PLAYLIST_PATH_LENGTH];
	snprintf(path, USER_PATH_LENGTH, "data/%s", g_logged->username);

	DIR* dir = opendir(path);
	struct dirent* entry;

	while ((entry = readdir(dir)) != NULL)
	{
		if (!strncmp(entry->d_name, "info", 4) || entry->d_name[0] == '.') continue;

		playlist* pl = malloc(sizeof(playlist));

		snprintf(fpath, PLAYLIST_PATH_LENGTH, "%s/%s", path, entry->d_name);

		FILE* f = fopen(fpath, "rb");
		fread(pl, sizeof(playlist), 1, f);
		fclose(f);

		printf("\nPlaylist: %s\n\"%s\"\n", pl->name, pl->description);
		for (uint8_t i = 0; i < pl->saved_songs; ++i)
			printf("\tSong %hhd: %s\n", i, pl->songs[i]);

		free(pl);
		pl = NULL;
	}

	closedir(dir);
}/*}}}*/

void play_random_song()
{/*{{{*/
	putchar('\n');
	switch (rand() % 5)
	{
		case 0:
			printf(u8"👟🦈👟 Tralalero Tralala 👟🦈👟\n");
			break;
		case 1:
			printf(u8"🌵🐘🌵 Lirili Larila 🌵🐘🌵\n");
			break;
		case 2:
			printf(u8"💣🐊💣 Bombardilo Crocodilo 💣🐊💣\n");
			break;
		case 3:
			printf(u8"🍂🌳🍂 Brr Brr Patapim 🍂🌳🍂\n");
			break;
		case 4:
			printf(u8"🐱🐟🐱 Brr Brr Patapim 🐱🐟🐱\n");
			break;
		default:
			break;
	}
}/*}}}*/

// vulns:
// can login with just the prefix of a password
// weak password generator
