# Documentación de la API REST - Sistema de Gestión de Inventario

Esta API REST implementa un sistema completo de gestión de inventario en Go bajo una **Arquitectura Hexagonal (Ports & Adapters)** con persistencia en **PostgreSQL**.

- **Prefijo base de todas las rutas:** `/api`
- **Formato de datos:** JSON (`application/json; charset=utf-8`)
- **Archivos CSV:** Codificación UTF-8 con soporte de BOM para Microsoft Excel y delimitadores `,` y `;`.

---

## 1. Módulos y Rutas Disponibles

### 1.1 Categorías (`/api/categorias`)

| Método | Endpoint | Descripción | Códigos HTTP |
|---|---|---|---|
| `GET` | `/api/categorias` | Listar todas las categorías (admite `?buscar=`) | `200` |
| `GET` | `/api/categorias/{id}` | Obtener una categoría por ID | `200`, `400`, `404` |
| `POST` | `/api/categorias` | Crear una categoría | `201`, `400`, `409`, `422` |
| `PATCH` | `/api/categorias/{id}` | Actualizar parcialmente una categoría | `200`, `400`, `404`, `409`, `422` |
| `DELETE` | `/api/categorias/{id}` | Eliminar una categoría | `204`, `400`, `404` |

#### Validaciones:
- `nombre`: Obligatorio, longitud máxima 100 caracteres, único en la base de datos.
- `color`: Formato hexadecimal `^#[0-9A-Fa-f]{6}$` (por defecto `#808080`).
- Si se elimina una categoría, sus productos y combos asociados quedan con `categoria_id = NULL` respetando la integridad referencial (`ON DELETE SET NULL`).

---

### 1.2 Proveedores (`/api/proveedores`)

| Método | Endpoint | Descripción | Códigos HTTP |
|---|---|---|---|
| `GET` | `/api/proveedores` | Listar todos los proveedores (admite `?buscar=`) | `200` |
| `GET` | `/api/proveedores/{id}` | Obtener un proveedor por ID | `200`, `400`, `404` |
| `POST` | `/api/proveedores` | Crear un proveedor | `201`, `400`, `409`, `422` |
| `PATCH` | `/api/proveedores/{id}` | Actualizar un proveedor | `200`, `400`, `404`, `409`, `422` |
| `DELETE` | `/api/proveedores/{id}` | Eliminar un proveedor | `204`, `400`, `404` |
| `GET` | `/api/proveedores/{id}/productos` | Consultar los productos asociados a este proveedor | `200`, `400`, `404` |

#### Validaciones:
- `nombre`: Obligatorio, longitud máxima 150 caracteres, único en la base de datos.
- Si se elimina un proveedor, los productos conservan su registro y pasan a tener `proveedor_id = NULL` (`ON DELETE SET NULL`).

---

### 1.3 Productos (`/api/productos`)

| Método | Endpoint | Descripción | Códigos HTTP |
|---|---|---|---|
| `GET` | `/api/productos` | Listar productos con filtros combinables, ordenamiento y paginación | `200` |
| `GET` | `/api/productos/{id}` | Obtener producto por ID | `200`, `400`, `404` |
| `POST` | `/api/productos` | Crear un producto | `201`, `400`, `409`, `422` |
| `PATCH` | `/api/productos/{id}` | Actualizar un producto | `200`, `400`, `404`, `409`, `422` |
| `DELETE` | `/api/productos/{id}` | Eliminar un producto | `204`, `400`, `404`, `409` |

#### Parámetros de consulta para `GET /api/productos`:
- `categoria_id` (int): Filtro por ID de categoría.
- `proveedor_id` (int): Filtro por ID de proveedor.
- `estado` (string): `activo` o `inactivo`.
- `buscar` (string): Busca coincidencias por nombre o código.
- `vencimiento_desde` / `vencimiento_hasta` (string, `AAAA-MM-DD`): Rango de fechas de vencimiento.
- `stock_min` / `stock_max` (float): Rango de stock.
- `precio_min` / `precio_max` (float): Rango de precios de venta.
- `stock_bajo` (boolean, `true`): Filtra productos cuyo stock sea inferior o igual al umbral.
- `umbral_stock_bajo` (float, por defecto `5`): Umbral configurable de stock bajo.
- `orden_por` (string): `nombre`, `codigo`, `precio_venta`, `costo`, `stock`, `vencimiento`, `creado_en`.
- `direccion` (string): `ASC` o `DESC`.
- `pagina` (int, default `1`) y `limite` (int, default `20`, máx `200`).

#### Unidades permitidas y conversiones:
- `unidad`: Unidad base de referencia (`unitario`, `kilogramo`, `gramo`, `litro`, `mililitro`).
- `costo_unidad`: Unidad en la que se expresa el `costo`.
- `venta_unidad`: Unidad en la que se expresa el `precio_venta`.
- El sistema realiza conversiones automáticas entre magnitudes compatibles (ej. 1 kg = 1000 g, 1 l = 1000 ml) para calcular con exactitud el costo unitario por medida base.

---

### 1.4 Combos (`/api/combos`)

| Método | Endpoint | Descripción | Códigos HTTP |
|---|---|---|---|
| `GET` | `/api/combos` | Listar combos (`?categoria_id=`, `?estado=`, `?buscar=`, `?pagina=`, `?limite=`) | `200` |
| `GET` | `/api/combos/{id}` | Consultar detalle de un combo con sus componentes, cantidades y costos | `200`, `400`, `404` |
| `POST` | `/api/combos` | Crear un combo (con componentes opcionales) | `201`, `400`, `409`, `422` |
| `PATCH` | `/api/combos/{id}` | Actualizar datos generales del combo | `200`, `400`, `404`, `409`, `422` |
| `DELETE` | `/api/combos/{id}` | Eliminar un combo (elimina sus componentes en cascada) | `204`, `400`, `404` |
| `POST` | `/api/combos/{id}/productos` | Agregar un producto componente al combo | `200`, `400`, `404`, `409`, `422` |
| `PATCH` | `/api/combos/{id}/productos/{producto_id}` | Modificar la cantidad de un producto componente | `200`, `400`, `404`, `422` |
| `DELETE` | `/api/combos/{id}/productos/{producto_id}` | Quitar un producto componente del combo | `200`, `400`, `404` |

#### Política de cálculo y sincronización de costos de Combos:
- **Cálculo dinámico y persistido:** El costo del combo se calcula sumando el costo proporcional de cada producto componente según su unidad de medida y cantidad:
  $$\text{Costo Componente} = \text{Cantidad} \times \text{Costo por Unidad Base del Producto}$$
- **Sincronización automática:**
  1. Al agregar, modificar o eliminar componentes de un combo, el costo en la tabla `combos` se recalcula y persiste inmediatamente.
  2. Si el costo de adquisición de un producto es modificado (mediante `PATCH /api/productos/{id}`), el sistema localiza todos los combos que contienen dicho producto y actualiza automáticamente sus costos almacenados.
- **Precio de venta independiente:** El `precio_venta` del combo se establece manualmente para permitir promociones comerciales y no depende de la suma de los precios de venta individuales.

---

### 1.5 Actualización Masiva de Precios

| Método | Endpoint | Descripción | Códigos HTTP |
|---|---|---|---|
| `POST` | `/api/productos/precios/previsualizar` | Simula la actualización sin modificar la base de datos | `200`, `400`, `422` |
| `POST` | `/api/productos/precios/actualizar` | Aplica la actualización porcentual efectiva en una transacción | `200`, `400`, `422` |

#### Cuerpo de solicitud (JSON):
```json
{
  "operacion": "aumento",
  "porcentaje": 5.5,
  "alcance": "categoria",
  "categoria_id": 2
}
```

- `operacion`: `"aumento"` o `"disminucion"` (o `"descuento"`).
- `porcentaje`: Número positivo mayor a cero (máximo configurable: 500%).
- `alcance`: `"todos"`, `"categoria"` (requiere `categoria_id`), `"proveedor"` (requiere `proveedor_id`).
- Redondeo: Los precios se calculan con 2 decimales mediante `ROUND(precio_venta * factor, 2)`.
- Previsualización: Devuelve `total_afectados` y la lista de productos con `precio_actual`, `nuevo_precio` y `diferencia`.
- **Política respecto a Combos:** La actualización masiva de precios solo altera el precio de venta de los productos. No modifica costos de productos, por lo que los costos de los combos no se alteran, ni se modifican automáticamente los precios de venta de los combos.

---

### 1.6 Importación y Exportación CSV

| Método | Endpoint | Descripción | Códigos HTTP |
|---|---|---|---|
| `POST` | `/api/productos/importar-csv` | Importar productos desde archivo CSV | `200`, `400`, `413`, `415`, `422` |
| `GET` | `/api/productos/exportar-csv` | Exportar productos a archivo CSV descargable | `200` |

#### Parámetros para `POST /api/productos/importar-csv`:
- Parámetros query:
  - `modo`: `"crear"` (por defecto, solo inserta; rechaza si el código ya existe) o `"actualizar"` (actualiza existentes, crea nuevos).
  - `previsualizar`: `"true"` o `"false"` (si es true, valida y cuenta sin persistir).
- Modos de envío:
  1. `multipart/form-data` con campo `"archivo"` o `"file"`.
  2. `text/csv` o cuerpo plano con las filas CSV directamente.
- Delimitadores admitidos: Detecta automáticamente tanto coma (`,`) como punto y coma (`;`).
- Validación por fila: Si existen errores de datos, formato o referencias, devuelve `422` con el detalle `{fila, codigo, campo, mensaje}` y ejecuta rollback transaccional para evitar importaciones corruptas.

#### Parámetros para `GET /api/productos/exportar-csv`:
- Admite los mismos filtros que el listado de productos (`categoria_id`, `proveedor_id`, `estado`, `buscar`, `stock_min`, `stock_max`, `precio_min`, `precio_max`, `stock_bajo`, etc.).
- Devuelve encabezado UTF-8 BOM (`\xEF\xBB\xBF`) y `Content-Type: text/csv; charset=utf-8` para compatibilidad nativa con Microsoft Excel en español.
