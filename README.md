# Solucionador de Sudokus en Go (Concurrente)

Este es un proyecto en **Go** que permite resolver Sudokus a partir de imágenes (capturas de pantalla) utilizando Visión por Computadora, OCR (Reconocimiento Óptico de Caracteres) y un algoritmo clásico de *Backtracking*.

Lo más destacado de este proyecto es su velocidad: utiliza **Goroutines** para paralelizar el procesamiento del OCR en las 81 celdas del tablero, logrando tiempos de ejecución casi instantáneos.

## Características Principales

*   **Visión por Computadora Ligera:** Utiliza un ingenioso sistema de proyecciones 1D para detectar automáticamente las líneas de la cuadrícula del Sudoku en la imagen.
*   **OCR Concurrente:** Se apoya en Tesseract OCR para leer los números, ejecutando el reconocimiento de todas las celdas de forma simultánea gracias a las rutinas de Go (`sync.WaitGroup`).
*   **Solver (Backtracking):** Algoritmo optimizado in-place para resolver cualquier Sudoku válido una vez extraída la información.
*   **Arquitectura Limpia:** Estructura de código idiomática en Go dividida en paquetes independientes (`vision`, `ocr`, y `sudoku`).

## Requisitos

Para que este programa funcione correctamente, necesitas:

1.  **Go** (versión 1.16 o superior).
2.  **Tesseract OCR** instalado en tu sistema.
    *   *En Windows:* Descarga e instala desde [UB Mannheim](https://github.com/UB-Mannheim/tesseract/wiki). Por defecto se busca en `C:\Program Files\Tesseract-OCR\tesseract.exe`.
    *   *En Linux:* `sudo apt install tesseract-ocr`
    *   *En macOS:* `brew install tesseract`

## Instalación y Compilación

1.  Clona este repositorio.
2.  Abre una terminal en la raíz del proyecto.
3.  Descarga las dependencias (principalmente `disintegration/imaging`):
    ```bash
    go mod download
    ```
4.  Compila el binario:
    ```bash
    go build -o sudokusolver.exe ./cmd/sudokusolver
    ```

## Uso

Una vez compilado, puedes pasarle como argumento la imagen de un Sudoku que quieras resolver usando el flag `-input`.

```bash
./sudokusolver.exe -input tu_imagen_sudoku.png
```

Opcionalmente, puedes usar el flag `-debug` para guardar una copia de la imagen que te mostrará cómo el programa detectó las líneas de la cuadrícula:

```bash
./sudokusolver.exe -input tu_imagen_sudoku.png -debug lineas_detectadas.png
```

## Estructura del Código

*   `/cmd/sudokusolver/main.go`: Punto de entrada de la aplicación. Orquesta la visión, el OCR paralelo y la resolución.
*   `/internal/vision`: Lógica de recorte de la imagen y detección de líneas de cuadrícula mediante proyecciones de color.
*   `/internal/ocr`: Interfaz para interactuar con el binario de Tesseract y preparar cada imagen individual.
*   `/internal/sudoku`: Lógica del juego y el algoritmo recursivo de Backtracking para la solución matemática.
