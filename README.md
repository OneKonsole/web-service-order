# One Konsole - Order Service
## Usage
docker build -t onekonsole/web-service-order:latest .
kind load docker-image onekonsole/web-service-order:latest
kubectl apply -f order-service-definition.yaml

## Sommaire
- Généralités
- Le système de MQ
    - Initialisation des files d'attente
    - Production de messages
    - Consommation des messages
        - Installation via HelmChart
        - Enregistrement de la commande
        - @TODO: Autres actions DB ?

- La base de données
- Le serveur HTTP
    - Création d'une commande
    - @TODO: Autres actions DB ? 

## Généralités
Nous avons décidé de créer un module Go permettant d'utiliser simplement et de façon généralisée les systèmes de créations de Queues ainsi que de consumers RabbitMQ. 

L'objectif ici est de pouvoir:
- lancer simplement une connection au service
- lancer simplement un consumer 

De ce fait, lorsque nous voulons utiliser des systèmes Rabbit, il suffit simplement d'appeler les méthodes de la bibliothèque. Cela permet de s'abstraire de la gestion du contexte Rabbit et de se concentrer sur la fonctionnalité à implémenter.   


Attention: 
Lorsque l'on veut que deux consumers utilisent le même message, nous devons configurer le auto-ack en true. Si nous le faisons manuellement, un des deux consumers pourrait ne pas recevoir le message.

