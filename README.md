# 📁 Cloud Storage

## Требования

Перед запуском убедитесь, что установлены:

- [Docker / Docker compose](https://docs.docker.com/desktop/)
- [Task](https://taskfile.dev/docs/installation)

## Установка

Склонируйте репозиторий и перейдите в него:

```
git clone --recurse-submodules https://github.com/Nurlan270/cloud-storage-go.git &&
 cd ./cloud-storage-go
```

## Запуск проекта

Для запуска проекта вы можете использовать один из вариантов ниже:

1. Запуск в локальной среде

    ```
    task app:start
    ```

2. Запуск в продакшене. Заменит `.env.example` на `.env` и выставит окружение на продовое,
   а также создаст продовую конфигурацию

    ```
    task app:deploy
    ```

> [!NOTE]
> Не забудьте добавить значения для переменных в `.env`.