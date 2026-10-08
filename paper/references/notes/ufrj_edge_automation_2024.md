# Sistema de Automacao de Infraestrutura para Aplicacoes de Edge Computing

**Autor:** REIS, Lucas de Queiroz dos
**Ano:** 2024
**Fonte:** Projeto de Graduacao (Eng. Eletronica e de Computacao), Escola Politecnica, UFRJ
**Orientador:** Prof. Luis Henrique Maciel Kosmalski Costa
**PDF:** `ufrj_edge_automation_2024.pdf`

## Resumo

Projeto de graduacao que desenvolve uma prova de conceito de pipeline CI/CD para integracao cloud-edge-IoT. Utiliza Jenkins para automacao CI/CD, Docker/Docker-compose para containerizacao, FastAPI + Django para camada de aplicacao, e NginX como reverse proxy/API gateway. O caso pratico e uma "ambulancia inteligente" com processamento de voz (SpeechRecognition, PyTorch, Librosa) e geolocalizacao. O sistema demonstra atualizacoes automatizadas de software em hardware embarcado na borda, sem intervencao manual, usando politicas de teste automatizado.

## Analise

Trabalho recente (2024) que aborda automacao de infraestrutura em contexto hibrido nuvem-borda. A fundamentacao teorica cobre DevOps, containers, orquestradores (KubeEdge, K3s), provisionamento de VMs, e monitoramento — conceitos que se sobrepoe ao seu TCC. A diferenca principal: este TCC foca em CI/CD para deployment de aplicacoes na borda, enquanto o seu foca em inicializacao automatica de instancias. Ambos preenchem lacunas onde ferramentas de grandes provedores nao estao disponiveis.

## Pontos-chave

- DevOps: metodologia que integra desenvolvimento e operacoes com foco em automacao
- CI/CD: automatiza build, teste e deployment com feedback rapido
- Compara orquestradores para edge: KubeEdge (baseado em Kubernetes para edge) vs K3s (Kubernetes leve)
- Provisionamento e configuracao: Terraform, Ansible, cloud-init mencionados como ferramentas fundamentais
- Jenkins como ferramenta de CI/CD para ambientes de borda (alternativa a GitLab CI/GitHub Actions)
- Pipeline inclui: build de imagens Docker -> testes automatizados -> deploy na borda
- Resultados: tempos de referencia estabelecidos para cada etapa do pipeline

## Informacoes Pertinentes para o TCC

- **Referencia recente (2024):** trabalho academico brasileiro atualizado na area de automacao de infraestrutura
- **Mencao a cloud-init:** o trabalho referencia cloud-init e provisionamento de VMs na fundamentacao teorica, validando a relevancia da tecnologia
- **Contexto hibrido nuvem-borda:** cenario onde metadata services sao ainda mais necessarios, pois infraestrutura local nao tem acesso a servicos de provedores cloud
- **Complementaridade:** seu metadata service poderia ser parte da camada de provisionamento que este tipo de pipeline CI/CD utiliza

## Sugestao BibTeX

```bibtex
@monography{reis2024sistema,
  title={Sistema de Automa{\c{c}}{\~a}o de Infraestrutura para Aplica{\c{c}}{\~o}es de Edge Computing},
  author={Reis, Lucas de Queiroz dos},
  year={2024},
  school={Escola Polit{\'e}cnica, Universidade Federal do Rio de Janeiro},
  address={Rio de Janeiro},
  type={Projeto de Gradua{\c{c}}{\~a}o (Eng. Eletr{\^o}nica e de Computa{\c{c}}{\~a}o)}
}
```
