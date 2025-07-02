# Запуск

- Необходим go версии 1.24.4

`go run cmd/main.go`

# Решение

## 1 - Ресерч

- Какие алгосы есть? -> Исследовал большинство популярных
- Какие были бы более эффективны? -> Арифметическое кодирование, Хаффман, семейство LZ
- Какие сложности? -> Рандомные данные, 50% сжатия +

Было решено начать с аримфетического кодирования, так как ранее не встречал этот алгоритм и было интересно попробовать. Меня очень зацепила концепция сжатия всех данных в одно единственное число

## 2 - ASCII, VLQ и Ints encoders

- ASCII encoder - до поиска более эффективного решения я быстро набросал схему, как компактно от 1 до 2 байт хранить любое число от 0 до 2^14. Оригинальное число представляется как 128 + 128 + ... + n, в первый байт пишется количество повторов 128 (если они есть), во второй байт пишется n - count\*128, то есть оставшаяся часть. У этого метода есть ограничение по размеру числа, поэтому я заменил его на более универсальный:
- VLQ encoder (Variable Length Quantity) - примерно то же самое, что придумал я, только расширенно для любого числа. Потенциально можно оптимизировать еще больше, так как на больших числах этот энкодер выдает не самые оптимальные результаты
- CompactVLQ encoder - я взял идею VLQ и соединил со своей и получил расширенный ASCII encoder, где число представляется как последовательность блоков, где первый бит показывает последний ли это блок, остальные предсталяют число от 0 до 2^7. Каждый последующий блок представляется как i \* 128^k, где i = [0:1], k - номер текущего блока, получается запись числа по основанию 128. Получается, что таким образом можно представить любое число от 0 до 2^7 как 1 байт, число от 2^7+1 до 2^14 как 2 байта, число от 2^14+1 до 2^21 как 3 байта и так далее
- IntsEncoder - не вписывается в концепцию. Это просто класс, который позволяет представить исходные данные []int в виде массива байт с помощью VLQEncoder

## 3 - Компрессоры

- ArithmeticCompessor - реализация алгоритма арифметического кодирования. Он требует слишком большой точности вычислений и выходные данные получались слишком большие. Так же подбор параметров для него напрямую зависит от размера данных и разнообразия символов - это чисто математическая задача (она супер интересная но я потратил на нее больше времени чем на решение самой задачи и не добился успеха). Было очень сложно подобрать необходимые параметры, чтобы заставить его эффективно работать, поэтому я отказался от этого решения
- HaffmanCompessor - реализация алгоритма Хаффмана
- LZ77Compessor - реализация алгоритма сжатия LZ77
- MixedCompessor = LZ77Compessor + HaffmanCompessor

## Анализ результатов

- Так как необязательно сохранять порядок входных данных, то было решено проверить результаты с сортировкой входных данных и без. Сортировка меняет игру в корне и дает LZ77 максимально раскрыть свой потенциал, так как появляются паттерны. Сжатие при отсортированных данных гораздо лучше
- При максимально равном распределении частот входных байт хорошо работает только LZ77, так как это его фишка. При таком распределении очень высокая вероятность появления паттернов
- Если разнообразие символов больше размера самого массива, то наблюдается сильная деградация всех алгоритмов. Если размер составляет более трети-четверти от разнообразия символов, то в целом алгоритмы работают хорошо. Иначе, чем больше разница между этими параметрами тем более разрастается архив вплоть до увеличения вдвое
- При полностью рандомных данных все алгоритмы работают одинаково плохо. Но алгоритм Хаффмана на удивление хорошо себя показывает именно здесь, если не сортировать входной массив. Скорее всего это связано с тем, что частоты символов примерно генерируются в +- нормальном распределении и здесь Хаффман показывает свою силу, а LZ77 работает плохо, так как нет абсолютно никаких паттернов
- Если большая часть сгенеренных данных это малое разнообразие повторяющихся символов, например, символы 1 и 2 составляют 96% выборки, то Хаффман работает очень эффективно
- Иногда mixed деградирует. Скорее всего это связанно с тем, что все данные так или иначе полностью основаны на рандоме. В среднем случае Хаффман после применения LZ77 дает прирост 1-4% в сжатии
- Алгоритмы на полностью рандомных данных при достаточно большом размере входных данных (10000 и более) дают очень неплохие результаты. Но Хаффман в одиночку не справляется, вероятно, из-за очень равномерного распределения частот символов
- Также все алгоритмы в целом очень эффективно работают при большом размере входного массива (10000 и более)

## Важная мысль

Так как я потратил уже и так слишком много времени на эту задачу, а эта мысль посетила меня только сейчас, я не буду ее реализовывать, но запишу. Входные данные можно отсортировать таким образом, чтобы они содержали как можно больше одинаковых паттернов. В таком случае, теоретически, в среднем можно добиться 80-90%+ сжатия только LZ77. Возможно LZ77 + Haffman дадут числа еще больше

## 1 - Сжатие при отсортированных входных данных

- Все алгоритмы были запущены 250 раз для высчитывания среднего
- `distributed randomly` значит полный рандом
- `where 96% of values are from 1 to 2, others are random` значит, что 96% входных данных это числа от 1 до 2, остальные от 1 до max, например, [1,2,1,1,1,2,1,2,1,256,123,11] - это попытка искуственно создать паттерны в абсолютном хаосе

| Haffman      |     LZ77     |    Mixed     | Test description                                                                                 |
| :----------- | :----------: | :----------: | :----------------------------------------------------------------------------------------------- |
| 70.496000%   |  58.384000%  |  49.688000%  | size = 50, with values from 1 to 50, where 96% of values are from 1 to 2, others are random      |
| 51.240000%   |  22.600000%  |  38.272000%  | size = 50, with values from 1 to 100, where 96% of values are from 1 to 4, others are random     |
| -2.761176%   | -44.173792%  | -19.514757%  | size = 50, with values from 1 to 300, where 96% of values are from 1 to 12, others are random    |
| 74.440000%   |  70.552000%  |  66.004000%  | size = 100, with values from 1 to 50, where 96% of values are from 1 to 2, others are random     |
| 57.748000%   |  46.864000%  |  53.828000%  | size = 100, with values from 1 to 100, where 96% of values are from 1 to 4, others are random    |
| 20.505566%   |  -7.029147%  |  20.629865%  | size = 100, with values from 1 to 300, where 96% of values are from 1 to 12, others are random   |
| 77.807200%   |  84.784800%  |  82.762400%  | size = 500, with values from 1 to 50, where 96% of values are from 1 to 2, others are random     |
| 63.574400%   |  76.696000%  |  77.664000%  | size = 500, with values from 1 to 100, where 96% of values are from 1 to 4, others are random    |
| 38.699938%   |  53.564183%  |  59.787805%  | size = 500, with values from 1 to 300, where 96% of values are from 1 to 12, others are random   |
| 79.155600%   |  88.021200%  |  86.724000%  | size = 1000, with values from 1 to 50, where 96% of values are from 1 to 2, others are random    |
| 64.994400%   |  82.937200%  |  82.613600%  | size = 1000, with values from 1 to 100, where 96% of values are from 1 to 4, others are random   |
| 40.986571%   |  66.756944%  |  68.762278%  | size = 1000, with values from 1 to 300, where 96% of values are from 1 to 12, others are random  |
| 83.767360%   |  94.909280%  |  95.502560%  | size = 10000, with values from 1 to 50, where 96% of values are from 1 to 2, others are random   |
| 69.863240%   |  92.961960%  |  93.262040%  | size = 10000, with values from 1 to 100, where 96% of values are from 1 to 4, others are random  |
| 46.085433%   |  86.642578%  |  86.281212%  | size = 10000, with values from 1 to 300, where 96% of values are from 1 to 12, others are random |
| 62.136000%   |  46.576000%  |  43.736000%  | size = 50, with values from 1 to 10, where 90% of values are from 1 to 2, others are random      |
| 38.528000%   |  8.656000%   |  20.168000%  | size = 50, with values from 1 to 50, where 90% of values are from 1 to 4, others are random      |
| 1.512000%    | -36.632000%  | -11.728000%  | size = 50, with values from 1 to 100, where 90% of values are from 1 to 9, others are random     |
| -73.255102%  | -101.404318% | -105.978260% | size = 50, with values from 1 to 300, where 90% of values are from 1 to 29, others are random    |
| 69.276000%   |  62.104000%  |  61.560000%  | size = 100, with values from 1 to 10, where 90% of values are from 1 to 2, others are random     |
| 45.260000%   |  34.036000%  |  39.040000%  | size = 100, with values from 1 to 50, where 90% of values are from 1 to 4, others are random     |
| 20.048000%   |  -1.148000%  |  19.464000%  | size = 100, with values from 1 to 100, where 90% of values are from 1 to 9, others are random    |
| -37.291814%  | -63.728911%  | -49.883130%  | size = 100, with values from 1 to 300, where 90% of values are from 1 to 29, others are random   |
| 79.780000%   |  83.582400%  |  86.568800%  | size = 500, with values from 1 to 10, where 90% of values are from 1 to 2, others are random     |
| 55.088000%   |  67.292800%  |  69.217600%  | size = 500, with values from 1 to 50, where 90% of values are from 1 to 4, others are random     |
| 35.836800%   |  52.453600%  |  55.011200%  | size = 500, with values from 1 to 100, where 90% of values are from 1 to 9, others are random    |
| 3.135396%    |  11.145187%  |  21.158070%  | size = 500, with values from 1 to 300, where 90% of values are from 1 to 29, others are random   |
| 81.955600%   |  89.265600%  |  91.650000%  | size = 1000, with values from 1 to 10, where 90% of values are from 1 to 2, others are random    |
| 59.642400%   |  75.942400%  |  77.816800%  | size = 1000, with values from 1 to 50, where 90% of values are from 1 to 4, others are random    |
| 39.803600%   |  65.544400%  |  67.086800%  | size = 1000, with values from 1 to 100, where 90% of values are from 1 to 9, others are random   |
| 9.798621%    |  33.921671%  |  38.957640%  | size = 1000, with values from 1 to 300, where 90% of values are from 1 to 29, others are random  |
| 83.937400%   |  97.404640%  |  97.729640%  | size = 10000, with values from 1 to 10, where 90% of values are from 1 to 2, others are random   |
| 67.466600%   |  93.017920%  |  94.026920%  | size = 10000, with values from 1 to 50, where 90% of values are from 1 to 4, others are random   |
| 50.458680%   |  89.114120%  |  90.439960%  | size = 10000, with values from 1 to 100, where 90% of values are from 1 to 9, others are random  |
| 25.760205%   |  76.871466%  |  79.526280%  | size = 10000, with values from 1 to 300, where 90% of values are from 1 to 29, others are random |
| 37.096000%   |  5.488000%   |  23.552000%  | size = 50, with values from 1 to 10, where 70% of values are from 1 to 3, others are random      |
| -49.656000%  | -83.432000%  | -73.240000%  | size = 50, with values from 1 to 50, where 70% of values are from 1 to 15, others are random     |
| -93.520000%  | -116.504000% | -125.248000% | size = 50, with values from 1 to 100, where 70% of values are from 1 to 30, others are random    |
| -133.606067% | -146.274621% | -177.593986% | size = 50, with values from 1 to 300, where 70% of values are from 1 to 90, others are random    |
| 51.488000%   |  33.112000%  |  48.452000%  | size = 100, with values from 1 to 10, where 70% of values are from 1 to 3, others are random     |
| -20.424000%  | -46.544000%  | -29.728000%  | size = 100, with values from 1 to 50, where 70% of values are from 1 to 15, others are random    |
| -60.236000%  | -83.300000%  | -78.388000%  | size = 100, with values from 1 to 100, where 70% of values are from 1 to 30, others are random   |
| -114.550856% | -120.205716% | -146.582844% | size = 100, with values from 1 to 300, where 70% of values are from 1 to 90, others are random   |
| 67.135200%   |  75.544000%  |  82.387200%  | size = 500, with values from 1 to 10, where 70% of values are from 1 to 3, others are random     |
| 20.517600%   |  30.280000%  |  44.324000%  | size = 500, with values from 1 to 50, where 70% of values are from 1 to 15, others are random    |
| -6.048000%   |  -0.099200%  |  11.102400%  | size = 500, with values from 1 to 100, where 70% of values are from 1 to 30, others are random   |
| -50.281752%  | -52.400492%  | -47.346558%  | size = 500, with values from 1 to 300, where 70% of values are from 1 to 90, others are random   |
| 69.356000%   |  84.860000%  |  88.714000%  | size = 1000, with values from 1 to 10, where 70% of values are from 1 to 3, others are random    |
| 29.474000%   |  53.266000%  |  64.686000%  | size = 1000, with values from 1 to 50, where 70% of values are from 1 to 15, others are random   |
| 7.596000%    |  29.698000%  |  40.154800%  | size = 1000, with values from 1 to 100, where 70% of values are from 1 to 30, others are random  |
| -28.162147%  | -19.997262%  | -10.160655%  | size = 1000, with values from 1 to 300, where 70% of values are from 1 to 90, others are random  |
| 71.221000%   |  97.070000%  |  97.417160%  | size = 10000, with values from 1 to 10, where 70% of values are from 1 to 3, others are random   |
| 37.942440%   |  90.212800%  |  92.313320%  | size = 10000, with values from 1 to 50, where 70% of values are from 1 to 15, others are random  |
| 23.875680%   |  83.736040%  |  87.061600%  | size = 10000, with values from 1 to 100, where 70% of values are from 1 to 30, others are random |
| 7.525928%    |  65.025136%  |  71.711369%  | size = 10000, with values from 1 to 300, where 70% of values are from 1 to 90, others are random |
| 18.472000%   | -25.520000%  |  10.384000%  | size = 50, with values from 1 to 10, distributed randomly                                        |
| -93.504000%  | -114.248000% | -124.200000% | size = 50, with values from 1 to 50, distributed randomly                                        |
| -130.392000% | -146.840000% | -169.384000% | size = 50, with values from 1 to 100, distributed randomly                                       |
| -114.391457% | -126.681857% | -156.327613% | size = 50, with values from 1 to 300, distributed randomly                                       |
| 39.240000%   |  14.164000%  |  45.716000%  | size = 100, with values from 1 to 10, distributed randomly                                       |
| -54.456000%  | -79.352000%  | -69.052000%  | size = 100, with values from 1 to 50, distributed randomly                                       |
| -103.556000% | -115.124000% | -129.036000% | size = 100, with values from 1 to 100, distributed randomly                                      |
| -101.740866% | -111.906759% | -134.689287% | size = 100, with values from 1 to 300, distributed randomly                                      |
| 55.888800%   |  71.514400%  |  79.350400%  | size = 500, with values from 1 to 10, distributed randomly                                       |
| 10.010400%   |  12.068800%  |  34.807200%  | size = 500, with values from 1 to 50, distributed randomly                                       |
| -21.441600%  | -26.367200%  |  -7.123200%  | size = 500, with values from 1 to 100, distributed randomly                                      |
| -44.473613%  | -56.491584%  | -45.682935%  | size = 500, with values from 1 to 300, distributed randomly                                      |
| 57.856000%   |  83.334400%  |  87.834400%  | size = 1000, with values from 1 to 10, distributed randomly                                      |
| 19.661600%   |  41.952400%  |  58.482400%  | size = 1000, with values from 1 to 50, distributed randomly                                      |
| -2.510400%   |  11.776000%  |  30.187200%  | size = 1000, with values from 1 to 100, distributed randomly                                     |
| -16.624628%  | -21.560588%  |  -1.499196%  | size = 1000, with values from 1 to 300, distributed randomly                                     |
| 59.510120%   |  96.891120%  |  97.275680%  | size = 10000, with values from 1 to 10, distributed randomly                                     |
| 27.618440%   |  89.679520%  |  91.865680%  | size = 10000, with values from 1 to 50, distributed randomly                                     |
| 14.558680%   |  82.129120%  |  86.534360%  | size = 10000, with values from 1 to 100, distributed randomly                                    |
| 17.692609%   |  68.718850%  |  75.453580%  | size = 10000, with values from 1 to 300, distributed randomly                                    |

## 2 - Сжатие без сортировки входных данных

| Haffman      |     LZ77     |    Mixed     | Test description                                                                                 |
| :----------- | :----------: | :----------: | :----------------------------------------------------------------------------------------------- |
| 70.200000%   |  58.060000%  |  48.940000%  | size = 50, with values from 1 to 50, where 96% of values are from 1 to 2, others are random      |
| 51.200000%   |  -7.520000%  |  -2.320000%  | size = 50, with values from 1 to 100, where 96% of values are from 1 to 4, others are random     |
| -2.639246%   | -67.065732%  | -56.100709%  | size = 50, with values from 1 to 300, where 96% of values are from 1 to 12, others are random    |
| 74.280000%   |  70.240000%  |  65.540000%  | size = 100, with values from 1 to 50, where 96% of values are from 1 to 2, others are random     |
| 57.720000%   |  7.540000%   |  9.110000%   | size = 100, with values from 1 to 100, where 96% of values are from 1 to 4, others are random    |
| 20.628876%   | -48.478936%  | -34.229950%  | size = 100, with values from 1 to 300, where 96% of values are from 1 to 12, others are random   |
| 77.824000%   |  84.648000%  |  81.866000%  | size = 500, with values from 1 to 50, where 96% of values are from 1 to 2, others are random     |
| 63.508000%   |  26.754000%  |  33.232000%  | size = 500, with values from 1 to 100, where 96% of values are from 1 to 4, others are random    |
| 38.647694%   | -21.516605%  |  -1.264158%  | size = 500, with values from 1 to 300, where 96% of values are from 1 to 12, others are random   |
| 79.117000%   |  87.612000%  |  85.552000%  | size = 1000, with values from 1 to 50, where 96% of values are from 1 to 2, others are random    |
| 64.942000%   |  31.025000%  |  42.647000%  | size = 1000, with values from 1 to 100, where 96% of values are from 1 to 4, others are random   |
| 41.055751%   | -16.572373%  |  8.484391%   | size = 1000, with values from 1 to 300, where 96% of values are from 1 to 12, others are random  |
| 83.767300%   |  92.443200%  |  93.032700%  | size = 10000, with values from 1 to 50, where 96% of values are from 1 to 2, others are random   |
| 69.865300%   |  36.697400%  |  56.690100%  | size = 10000, with values from 1 to 100, where 96% of values are from 1 to 4, others are random  |
| 46.108122%   | -10.470305%  |  23.126314%  | size = 10000, with values from 1 to 300, where 96% of values are from 1 to 12, others are random |
| 62.980000%   |  45.820000%  |  40.820000%  | size = 50, with values from 1 to 10, where 90% of values are from 1 to 2, others are random      |
| 38.440000%   | -20.360000%  | -13.620000%  | size = 50, with values from 1 to 50, where 90% of values are from 1 to 4, others are random      |
| 2.000000%    | -61.160000%  | -51.080000%  | size = 50, with values from 1 to 100, where 90% of values are from 1 to 9, others are random     |
| -73.583387%  | -109.912489% | -128.616804% | size = 50, with values from 1 to 300, where 90% of values are from 1 to 29, others are random    |
| 69.280000%   |  60.490000%  |  58.980000%  | size = 100, with values from 1 to 10, where 90% of values are from 1 to 2, others are random     |
| 44.990000%   |  -3.110000%  |  -0.080000%  | size = 100, with values from 1 to 50, where 90% of values are from 1 to 4, others are random     |
| 20.050000%   | -43.910000%  | -32.790000%  | size = 100, with values from 1 to 100, where 90% of values are from 1 to 9, others are random    |
| -37.655437%  | -85.005876%  | -88.446829%  | size = 100, with values from 1 to 300, where 90% of values are from 1 to 29, others are random   |
| 79.800000%   |  79.878000%  |  81.670000%  | size = 500, with values from 1 to 10, where 90% of values are from 1 to 2, others are random     |
| 55.106000%   |  19.504000%  |  30.878000%  | size = 500, with values from 1 to 50, where 90% of values are from 1 to 4, others are random     |
| 35.660000%   | -17.080000%  |  0.422000%   | size = 500, with values from 1 to 100, where 90% of values are from 1 to 9, others are random    |
| 3.125603%    | -54.867879%  | -37.269024%  | size = 500, with values from 1 to 300, where 90% of values are from 1 to 29, others are random   |
| 81.979000%   |  83.811000%  |  85.111000%  | size = 1000, with values from 1 to 10, where 90% of values are from 1 to 2, others are random    |
| 59.634000%   |  25.490000%  |  41.534000%  | size = 1000, with values from 1 to 50, where 90% of values are from 1 to 4, others are random    |
| 39.985000%   | -10.507000%  |  12.823000%  | size = 1000, with values from 1 to 100, where 90% of values are from 1 to 9, others are random   |
| 9.642191%    | -48.631633%  | -25.301663%  | size = 1000, with values from 1 to 300, where 90% of values are from 1 to 29, others are random  |
| 83.934000%   |  88.626700%  |  90.792200%  | size = 10000, with values from 1 to 10, where 90% of values are from 1 to 2, others are random   |
| 67.454800%   |  32.186500%  |  54.140400%  | size = 10000, with values from 1 to 50, where 90% of values are from 1 to 4, others are random   |
| 50.457700%   |  -2.146900%  |  29.139500%  | size = 10000, with values from 1 to 100, where 90% of values are from 1 to 9, others are random  |
| 25.761833%   | -39.785775%  |  -3.252751%  | size = 10000, with values from 1 to 300, where 90% of values are from 1 to 29, others are random |
| 37.840000%   | -13.640000%  |  0.700000%   | size = 50, with values from 1 to 10, where 70% of values are from 1 to 3, others are random      |
| -49.800000%  | -96.320000%  | -100.540000% | size = 50, with values from 1 to 50, where 70% of values are from 1 to 15, others are random     |
| -92.240000%  | -121.100000% | -147.040000% | size = 50, with values from 1 to 100, where 70% of values are from 1 to 30, others are random    |
| -134.426830% | -149.208043% | -202.450731% | size = 50, with values from 1 to 300, where 70% of values are from 1 to 90, others are random    |
| 51.590000%   |  3.250000%   |  17.390000%  | size = 100, with values from 1 to 10, where 70% of values are from 1 to 3, others are random     |
| -20.060000%  | -73.820000%  | -65.850000%  | size = 100, with values from 1 to 50, where 70% of values are from 1 to 15, others are random    |
| -59.980000%  | -96.830000%  | -108.930000% | size = 100, with values from 1 to 100, where 70% of values are from 1 to 30, others are random   |
| -113.473872% | -126.203273% | -173.522663% | size = 100, with values from 1 to 300, where 70% of values are from 1 to 90, others are random   |
| 67.194000%   |  29.840000%  |  40.410000%  | size = 500, with values from 1 to 10, where 70% of values are from 1 to 3, others are random     |
| 20.352000%   | -38.568000%  |  -9.644000%  | size = 500, with values from 1 to 50, where 70% of values are from 1 to 15, others are random    |
| -6.054000%   | -59.774000%  | -40.436000%  | size = 500, with values from 1 to 100, where 70% of values are from 1 to 30, others are random   |
| -50.177052%  | -84.229883%  | -90.572320%  | size = 500, with values from 1 to 300, where 70% of values are from 1 to 90, others are random   |
| 69.328000%   |  35.302000%  |  48.624000%  | size = 1000, with values from 1 to 10, where 70% of values are from 1 to 3, others are random    |
| 29.518000%   | -31.366000%  |  2.716000%   | size = 1000, with values from 1 to 50, where 70% of values are from 1 to 15, others are random   |
| 7.799000%    | -51.726000%  | -23.000000%  | size = 1000, with values from 1 to 100, where 70% of values are from 1 to 30, others are random  |
| -28.323049%  | -74.264107%  | -67.129343%  | size = 1000, with values from 1 to 300, where 70% of values are from 1 to 90, others are random  |
| 71.227000%   |  41.113000%  |  60.133000%  | size = 10000, with values from 1 to 10, where 70% of values are from 1 to 3, others are random   |
| 37.943200%   | -24.224000%  |  15.189300%  | size = 10000, with values from 1 to 50, where 70% of values are from 1 to 15, others are random  |
| 23.874500%   | -43.560900%  |  -4.663900%  | size = 10000, with values from 1 to 100, where 70% of values are from 1 to 30, others are random |
| 7.532708%    | -62.593136%  | -30.221632%  | size = 10000, with values from 1 to 300, where 70% of values are from 1 to 90, others are random |
| 18.860000%   | -54.020000%  | -34.860000%  | size = 50, with values from 1 to 10, distributed randomly                                        |
| -96.400000%  | -125.660000% | -154.880000% | size = 50, with values from 1 to 50, distributed randomly                                        |
| -129.380000% | -150.200000% | -194.020000% | size = 50, with values from 1 to 100, distributed randomly                                       |
| -116.582634% | -126.894834% | -190.690185% | size = 50, with values from 1 to 300, distributed randomly                                       |
| 39.270000%   | -33.740000%  | -14.650000%  | size = 100, with values from 1 to 10, distributed randomly                                       |
| -55.050000%  | -96.410000%  | -104.940000% | size = 100, with values from 1 to 50, distributed randomly                                       |
| -104.450000% | -124.670000% | -161.500000% | size = 100, with values from 1 to 100, distributed randomly                                      |
| -101.210422% | -110.503791% | -166.336128% | size = 100, with values from 1 to 300, distributed randomly                                      |
| 55.872000%   |  -7.298000%  |  16.034000%  | size = 500, with values from 1 to 10, distributed randomly                                       |
| 10.042000%   | -57.992000%  | -27.598000%  | size = 500, with values from 1 to 50, distributed randomly                                       |
| -21.510000%  | -76.864000%  | -60.002000%  | size = 500, with values from 1 to 100, distributed randomly                                      |
| -44.425509%  | -78.154581%  | -94.188125%  | size = 500, with values from 1 to 300, distributed randomly                                      |
| 57.832000%   |  -2.869000%  |  25.620000%  | size = 1000, with values from 1 to 10, distributed randomly                                      |
| 19.668000%   | -52.012000%  | -15.141000%  | size = 1000, with values from 1 to 50, distributed randomly                                      |
| -2.524000%   | -68.985000%  | -39.800000%  | size = 1000, with values from 1 to 100, distributed randomly                                     |
| -16.914063%  | -72.077680%  | -64.392321%  | size = 1000, with values from 1 to 300, distributed randomly                                     |
| 59.501300%   |  1.219300%   |  37.806700%  | size = 10000, with values from 1 to 10, distributed randomly                                     |
| 27.618300%   | -46.543600%  |  -3.400800%  | size = 10000, with values from 1 to 50, distributed randomly                                     |
| 14.563000%   | -62.056500%  | -21.622300%  | size = 10000, with values from 1 to 100, distributed randomly                                    |
| 17.682554%   | -65.650931%  | -28.987168%  | size = 10000, with values from 1 to 300, distributed randomly                                    |

## 3 - Compression results of edge cases

`distributed equally` значит [1,2,3,...,1,2,3,...,1,2,3,...]

| Haffman     |    LZ77     |    Mixed    | Test description                                           |
| :---------- | :---------: | :---------: | :--------------------------------------------------------- |
| -16.666667% | -20.000000% | -16.666667% | size = 30, with values from 1 to 10, distributed equally   |
| -40.666667% | -4.000000%  | -9.333333%  | size = 150, with values from 1 to 50, distributed equally  |
| -51.666667% | -2.000000%  | -90.333333% | size = 300, with values from 1 to 100, distributed equally |
| -23.819591% | 23.326286%  | -65.468640% | size = 900, with values from 1 to 300, distributed equally |

## 4 - Результат сжатия logfile.txt

- Файлик был сгенерен нейронкой, но близок к реальным данным
- Mixed отработал плохо. Почему - до конца не ясно, скорее всего это связано с тем, что в нем очень много паттернов, LZ77 убирает их практически полностью, оставляя для Хаффмана равномерно распределенные частоты, с чем он не справляется. Вероятно нужно встраивать анализаторы результата LZ77 прежде чем отдавать их на повторное сжатие черех Хаффмана

| Haffman    |    LZ77    |   Mixed    | Test description |
| :--------- | :--------: | :--------: | :--------------- |
| 39.233363% | 68.138884% | 51.557904% | logfile.txt      |
