CREATE DATABASE IF NOT EXISTS db_post;
USE db_post;

START TRANSACTION;

CREATE TABLE categorias (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    color VARCHAR(7) NOT NULL DEFAULT '#808080',
    CONSTRAINT chk_categorias_color CHECK (color REGEXP '^#[0-9A-Fa-f]{6}$')
);

CREATE TABLE proveedores (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(150) NOT NULL UNIQUE
);

CREATE TABLE productos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(150) NOT NULL,
    codigo VARCHAR(100) NOT NULL UNIQUE,
    descripcion TEXT,
    categoria_id BIGINT,
    unidad VARCHAR(20) NOT NULL DEFAULT 'unitario',
    costo_unidad VARCHAR(20) NOT NULL DEFAULT 'unitario',
    venta_unidad VARCHAR(20) NOT NULL DEFAULT 'unitario',
    costo DECIMAL(12, 2) NOT NULL DEFAULT 0,
    precio_venta DECIMAL(12, 2) NOT NULL DEFAULT 0,
    stock DECIMAL(12, 3) NOT NULL DEFAULT 0,
    proveedor_id BIGINT,
    vencimiento DATE,
    imagen TEXT,
    estado VARCHAR(20) NOT NULL DEFAULT 'activo',
    creado_en TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_productos_categoria FOREIGN KEY (categoria_id) REFERENCES categorias(id) ON DELETE SET NULL,
    CONSTRAINT fk_productos_proveedor FOREIGN KEY (proveedor_id) REFERENCES proveedores(id) ON DELETE SET NULL,
    CONSTRAINT chk_productos_unidad CHECK (unidad IN ('unitario', 'kilogramo', 'gramo', 'litro', 'mililitro')),
    CONSTRAINT chk_productos_costo_unidad CHECK (costo_unidad IN ('unitario', 'kilogramo', 'gramo', 'litro', 'mililitro')),
    CONSTRAINT chk_productos_venta_unidad CHECK (venta_unidad IN ('unitario', 'kilogramo', 'gramo', 'litro', 'mililitro')),
    CONSTRAINT chk_productos_costo CHECK (costo >= 0),
    CONSTRAINT chk_productos_precio_venta CHECK (precio_venta >= 0),
    CONSTRAINT chk_productos_stock CHECK (stock >= 0),
    CONSTRAINT chk_productos_estado CHECK (estado IN ('activo', 'inactivo'))
);

CREATE TABLE combos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(150) NOT NULL,
    codigo VARCHAR(100) NOT NULL UNIQUE,
    categoria_id BIGINT,
    imagen TEXT,
    estado VARCHAR(20) NOT NULL DEFAULT 'activo',
    costo DECIMAL(12, 2) NOT NULL DEFAULT 0,
    precio_venta DECIMAL(12, 2) NOT NULL DEFAULT 0,
    creado_en TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_combos_categoria FOREIGN KEY (categoria_id) REFERENCES categorias(id) ON DELETE SET NULL,
    CONSTRAINT chk_combos_estado CHECK (estado IN ('activo', 'inactivo')),
    CONSTRAINT chk_combos_costo CHECK (costo >= 0),
    CONSTRAINT chk_combos_precio_venta CHECK (precio_venta >= 0)
);

CREATE TABLE combo_productos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    combo_id BIGINT NOT NULL,
    producto_id BIGINT NOT NULL,
    cantidad DECIMAL(12, 3) NOT NULL DEFAULT 1,
    
    CONSTRAINT fk_combo_productos_combo FOREIGN KEY (combo_id) REFERENCES combos(id) ON DELETE CASCADE,
    CONSTRAINT fk_combo_productos_producto FOREIGN KEY (producto_id) REFERENCES productos(id) ON DELETE RESTRICT,
    CONSTRAINT chk_combo_productos_cantidad CHECK (cantidad > 0),
    CONSTRAINT uq_combo_producto UNIQUE (combo_id, producto_id)
);

CREATE INDEX idx_productos_nombre ON productos(nombre);
CREATE INDEX idx_productos_categoria ON productos(categoria_id);
CREATE INDEX idx_productos_proveedor ON productos(proveedor_id);
CREATE INDEX idx_productos_estado ON productos(estado);
CREATE INDEX idx_productos_vencimiento ON productos(vencimiento);
CREATE INDEX idx_combos_categoria ON combos(categoria_id);
CREATE INDEX idx_combos_estado ON combos(estado);
CREATE INDEX idx_combo_productos_producto ON combo_productos(producto_id);

COMMIT;