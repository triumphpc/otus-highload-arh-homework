Для выполнения задания поднял:
- кластер pg с patroni для кворумной синхронизации 
- экспортер метрик pg в графану 
- pgpool для распределении нагрузки между репликами

![img_1.png](img_1.png)

![img_2.png](img_2.png)



Нагенерил тестовых пользователей и проверил как работает асинхронная репликация на всех 3 реплик 

![img.png](img.png)


Настроил jmeter на стратегию нагрузочного тестирования и запустил тест на чтение двух ручек 
и на получения рандомных пользователей и поиск по имени фамилии. 

Далее погасил все реплики - оставил только master  и дал нагрузку на 60 sec:

![img_3.png](img_3.png)

![img_4.png](img_4.png)


Далее, включил реплики в асинхронном режиме и дал нагрузку на 60 sec (master-slave-slave):

![img_5.png](img_5.png)

Видно что нагрузка распределилась по всем репликам на чтение 

![img_6.png](img_6.png)
![img_7.png](img_7.png)
![img_8.png](img_8.png)

Видно, что обработали практически в два раза больше запросов. 

Далее в сценарий добавил запрос на запись и продлил тест на 2 минуты.  
После минуты работы остановил мастер и наблюдал 3 секундный простой на запись и возобновление записи во вторую реплику, после того, 
как patroni переключил его как master.

![img_9.png](img_9.png)

![img_10.png](img_10.png)

Вот тут видно, что в patroni1 был мастером и перестал писать 

![img_11.png](img_11.png)


а потом начал писать вторая реплика patroni2

![img_12.png](img_12.png)

Видно по графикам, что был простой транзакций и какие-то потери. Сравнил с ситуацией без простоя по количеству запросов 2 мин:

![img_13.png](img_13.png)

Видно, что простои в скорости обработки точно есть в два раза. Повлияло время переключения. 













