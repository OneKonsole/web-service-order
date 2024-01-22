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

Useful commands:
helm install web-order ./web-order-chart -f ./web-order-chart/values.yaml

kind load docker-image onekonsole/web-service-order:latest
docker build -t onekonsole/web-service-order:latest .

kubectl run postgresql-client --rm --tty -i --restart='Never' \
--namespace default \
--image docker.io/bitnami/postgresql:11.6.0-debian-10-r0 \
--env="PGPASSWORD=root" \
--command -- psql --host my-postgresql.provisioning \
-U root -d order -p 5432

kubectl run rabbitmq-client --rm --tty -i --restart='Never' \
--namespace provisioning \
--image docker.io/bitnami/rabbitmq:3.8.14-debian-10-r0 \
--env="RABBITMQ_PASSWORD=admin" \
--env="RABBITMQ_USERNAME=admin" \
--command -- rabbitmqctl list_queues --longnames -n rabbit@my-rabbitmq.provisioning.svc.cluster.local


export served_port=8010
export db_user=root
export db_password=root
export db_URL=my-postgresql.provisioning.svc.cluster.local
export db_name=order
export sys_service_url=http://sys-order.provisioning.svc.cluster.local:8020/produce/order