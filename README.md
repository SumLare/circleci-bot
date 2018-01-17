## CircleCI Bot

This is a telegram bot which helps send CircleCI notifications through telegram.

## Installation

Use /add command in chat to obtain the circleci key

Add following lines to your application's .circleci/config.yml:

```yml
notify:
  webhooks:
    - url: https://circle-ci-bot.herokuapp.com/hooks/circle?circleci_key=your_key
```
