package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Metricas struct {
	UsoCPU   float64 `json:"uso_cpu"`
	UsoDisco float64 `json:"uso_disco"`
}

// obtenerUsoDisco ejecuta el comando 'df' y parsea su salida para obtener el porcentaje de uso.
func obtenerUsoDisco() (float64, error) {
	// Comando: df -h --output=pcent /
	// Esto nos da el porcentaje de uso de la partición raíz (/)
	cmd := exec.Command("df", "-h", "--output=pcent", "/")
	
	// Ejecutamos el comando y capturamos su salida combinada (stdout y stderr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("error ejecutando df: %v, salida: %s", err, string(out))
	}

	// La salida será algo como:
	// Pcent
	//   15%
	// La dividimos por líneas
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("salida inesperada de df: %s", string(out))
	}

	// Tomamos la segunda línea, quitamos espacios y el símbolo '%'
	usoStr := strings.TrimSpace(lines[1])
	usoStr = strings.TrimSuffix(usoStr, "%")

	// Convertimos el string a un número flotante
	uso, err := strconv.ParseFloat(usoStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error convirtiendo el uso de disco a número: %v", err)
	}

	return uso, nil
}

func obtenerUsoCPU() (float64, error) {
	// Comando: top -bn1 | grep 'Cpu(s)' | awk '{print $2 + $4}'
	// - top -bn1: Ejecuta top en modo batch (b) una vez (n1).
	// - grep 'Cpu(s)': Filtra la línea que contiene la info de la CPU.
	// - awk '{print $2 + $4}': Suma el porcentaje de usuario ($2) y el de sistema ($4).
	// Usamos 'sh -c' para poder interpretar los pipes '|'
	cmdStr := "top -bn1 | grep 'Cpu(s)' | awk '{print $2 + $4}'"
	cmd := exec.Command("sh", "-c", cmdStr)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("error ejecutando top: %v, salida: %s", err, string(out))
	}
	
	// La salida es directamente el número que queremos, pero como string.
	usoStr := strings.TrimSpace(string(out))

	// Reemplazamos la coma decimal por un punto si es necesario (depende del idioma del SO)
	usoStr = strings.Replace(usoStr, ",", ".", -1)

	uso, err := strconv.ParseFloat(usoStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error convirtiendo el uso de cpu a número: %v", err)
	}

	return uso, nil
}


func main() {
	for {
		metricas := Metricas{}
		var errDisco, errCPU error

		metricas.UsoDisco, errDisco = obtenerUsoDisco()
		if errDisco != nil {
			// Usamos log para un formato de error más estándar
			log.Printf("Error obteniendo uso de disco: %v\n", errDisco)
		}

		metricas.UsoCPU, errCPU = obtenerUsoCPU()
		if errCPU != nil {
			log.Printf("Error obteniendo uso de CPU: %v\n", errCPU)
		}

		// Convertimos nuestro struct a un slice de bytes en formato JSON
		// MarshalIndent formatea el JSON para que sea legible por humanos (con sangría)
		jsonData, err := json.MarshalIndent(metricas, "", "  ")
		if err != nil {
			log.Printf("Error al convertir a JSON: %v\n", err)
			continue // Si no podemos crear el JSON, saltamos esta iteración
		}

		// Imprimimos el JSON en la consola
		fmt.Println("--- Nuevas Métricas ---")
		fmt.Println(string(jsonData))
		
		time.Sleep(10 * time.Second)
	}
}
