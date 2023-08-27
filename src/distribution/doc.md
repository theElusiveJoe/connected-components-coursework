функция FindDistribution принимает на вход:
* итератор ребер
* количество слейвов
* некоторый hashnum

и возвращает массив hashToSlave:
hashToSlave[h] = n <=> за ноды с хэшем h отвечает слейв с номером n.

# что происзодит под капотом?


## предварительно
Функция нахождения хэша принимает на вход 
id ноды и возвращает id%hashnum

Таким образом, логично было бы ставить hashnum не превосходящим количество нод N, т.к. отношение hashnum/N - есть коэффициент сжатия мощности множества вершин при переходе от изначального графа к "мультиграфу".

Здесь под "мультиграфом" подразумевается граф, полученный объединением нод с одинаковыми хэшами в одну мультиноду.

## этапы работы

### 0 инициализация мультиграфа
Каждая мультинода имеет вес - количество выходящих ребер из ее прообраза в изначальном графе (т.е. петля на мультиноде добавляет сразу 2 очка к ее весу, хотя и не будет отражена в мультиграфе).

### 1 нахождение связных мультикомпонент
Мультиграф перегоняется в обычный граф, на который действует обычный алгоритм нахождения мультикомпонент. 

Вычисляется вес мультикомпоненты - сумма весов мультинод в этой мультикомпоненте.

Сортируем мультикомпоненты по их весам.

### 2 распределение по слейвам

Сейчас мы имеем отсортированную последовательнось мультикомпонент. Формально, мультикомпонента M представляет из себя подмножество нод изначального графа такое, что:

1. если h(v1) = h1 и v1 in M => для любой v2: h(v2) = h1 => v2 in M (т.е. все ноды одного хеша лежат только в одной компоненте).
2. если есть ребро (v1, v2) и v1 in M => v2 in M.

Свойство 1 нам нужно было, чтобы ускорить нахождение мультикомпонент. Чем меньше можность множества значений хэш-функции, тем меньший размер имеет мультиграф.

Свойство 2 дает нам подсказку о том, как распределить ноды по слейвам, чтобы минимизировать общение между нодами.

создаем массив длинны slavesNum. элементы этого массива - множество хешей, за которые будет отвечать этот слейв.

И теперь начинаем "змейкой" распределять мультикомпоненты по слейвам. 

#### отступление
Мы хотим равномерно распределить нагрузку на слейвов. Логично было бы распределить ноды по слейвам так, чтобы минимизировать дисперсию количества нод на слейвах. Однако (при условии децентрализованности графа), при большом количестве нод, "вес" мультинод будет кореллировать с количеством нод в мультиноде. Поэтому, можно распределять мультиноды, опираясь на их вес. 

Также, считаем, что на каждый хэш приходится +- одинаковое количество нод (отображение %hashNum под это условие подходит).

#### продолжим

Распределяем мультикомпоненты по нодам. Каждую мультикомпоненту пытаемся максимально "впихнуть" одному слейву. Максимум, один слейв владеет 120% хешей от равномерного распределения. 

получили отображение номер слейва -> подконтрольное ему множество хешей.

потом это перегоняется в обратное отображение и возвращается в виде одномерного массива.