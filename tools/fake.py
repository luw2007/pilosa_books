# -*- coding: utf-8 -*-

import datetime
import random

MAX_USER_ID = 100000000
MAX_BOOK_ID = 100000
DAILY_USER_COUNT = 10000000

DEFAULT_DAYS = 7
DEFAULT_ACTIONS = ['read', 'down', 'shelf']


def init(days=DEFAULT_DAYS, actions=DEFAULT_ACTIONS):
    now = datetime.datetime.now()
    days_array = [(now -  datetime.timedelta(days=x)).strftime("%Y-%m-%dT00:00") for x in range(days)]

    users = set()
    for action in actions:
        with open("csv/{action}_day.csv".format(action=action), "w") as f:
            for day in days_array:
                for i in range(DAILY_USER_COUNT):
                    user = random.randint(0, MAX_USER_ID)
                    users.add(user)
                    f.write("1,{},{}\n".format(user, day))

        with open("csv/{action}_book_count.csv".format(action=action), "w") as f:
            for i in range(DAILY_USER_COUNT):
                user = random.randint(0, MAX_USER_ID)
                users.add(user)
                count = random.randint(1, 3)
                f.write("{},{}\n".format(count, user))

    with open("csv/gender.csv", "w") as f:
        for user in users:
            gender = random.randint(0, 2)
            f.write("{},{}\n".format(gender, user))

if __name__ == "__main__":
    try:
        import fire
        fire.Fire()
    except ImportError:
        init()
