# Cobro después de la atención

Esta página muestra cómo paga el paciente una consulta una vez que el profesional lo atendió.
El **cobro** es el paso entre «la atención terminó» y «la clínica tiene el dinero y el
paciente tiene su boleta» (el documento que el SII, el Servicio de Impuestos Internos, exige
por cada venta). Te lo encontrarás al tocar el cierre de una atención, `payment_manager` o
`cash_manager`.

La clínica tiene dos máquinas de cobro TUU (terminales de tarjeta de la empresa chilena
Haulmer, ver [veltylabs/tuu](https://github.com/veltylabs/tuu)): una para las consultas y otra
para la farmacia, así cada caja se cuadra por separado. La máquina imprime la boleta tanto
en efectivo como con tarjeta.

## 1. Hoy (papelito)

El profesional atiende en el segundo piso; la caja está en el primero.

```mermaid
flowchart TD
    A[El profesional termina la atención] --> B[Le entrega al paciente un papelito<br/>con el servicio]
    B --> C[El paciente baja a la caja]
    C --> D{¿Paga en efectivo o con tarjeta?}
    D -- efectivo --> E[La cajera digita el monto<br/>en la TUU y recibe el efectivo]
    D -- tarjeta --> F[La cajera digita el monto<br/>y el paciente pasa la tarjeta]
    E --> G[La TUU imprime la boleta<br/>autorizada por el SII]
    F --> G
    G --> H[El paciente vuelve a subir<br/>y muestra la boleta]
    H --> I[El profesional anota la atención<br/>en su registro personal]
    I --> J{¿El paciente pide<br/>boleta de honorarios?}
    J -- sí --> K[El profesional la emite<br/>por su cuenta]
    J -- no --> L[Fin]
    K --> L
```

## 2. Decidido (sin papel; la cajera solo digita los pagos en efectivo)

> STATUS (borrar esta nota cuando cierre MONEY_FLOWS_MASTER_PLAN): flujo decidido; qué procesador
> en línea se usa sigue abierto en ese master.

Una **solicitud de pago** es la cuenta de una consulta: el paciente, el profesional que la
envió, la hora y sus servicios (uno o varios), cada uno con la tarifa del servicio. El
profesional no puede cambiar el monto.

```mermaid
flowchart TD
    A[El profesional termina la atención<br/>con uno o más servicios] --> B{¿Cómo paga el paciente?}
    B -- en caja --> C[Aprieta Enviar a caja<br/>y cierra la atención<br/>sin esperar al paciente]
    C --> D[La cajera con turno abierto<br/>recibe el aviso]
    D --> E[Lista de pendientes: paciente,<br/>profesional, hora, monto]
    E --> F[La cajera pregunta el nombre<br/>y selecciona la solicitud]
    F --> G{Medio de pago}
    G -- tarjeta --> H[El sistema envía el monto<br/>a la máquina de la caja]
    H --> I[El paciente pasa la tarjeta<br/>la máquina imprime la boleta]
    G -- efectivo --> J[La cajera recibe el efectivo<br/>y lo digita en la TUU]
    J --> K[La TUU imprime la boleta<br/>y la cajera confirma en el sistema]
    B -- en línea ahora --> L[La pantalla muestra un QR<br/>con el enlace de pago]
    L --> M[El paciente paga en su teléfono]
    M --> N[El procesador avisa al servidor]
    I --> O[Solicitud marcada Pagada]
    K --> O
    N --> O
    O --> P[El profesional ve Pagada<br/>en esa consulta]
```

La máquina de cobro y el procesador en línea (TUU hoy, Transbank u otro mañana) se eligen con
una configuración en la base de datos, no en el código: los módulos solo ven los puertos de
`payment_manager`.
